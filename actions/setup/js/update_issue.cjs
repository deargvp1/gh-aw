// @ts-check
/// <reference types="@actions/github-script" />

/**
 * @typedef {import('./types/handler-factory').HandlerFactoryFunction} HandlerFactoryFunction
 */

/** @type {string} Safe output type handled by this module */
const HANDLER_TYPE = "update_issue";

const { resolveTarget } = require("./safe_output_helpers.cjs");
const { createUpdateHandlerFactory, createStandardResolveNumber, createStandardFormatResult } = require("./update_handler_factory.cjs");
const { updateBody } = require("./update_pr_description_helpers.cjs");
const { loadTemporaryProjectMap, replaceTemporaryProjectReferences } = require("./temporary_id.cjs");
const { sanitizeTitle } = require("./sanitize_title.cjs");
const { tryEnforceArrayLimit } = require("./limit_enforcement_helpers.cjs");
const { ERR_VALIDATION } = require("./error_codes.cjs");
const { parseBoolTemplatable } = require("./templatable.cjs");
const { buildWorkflowRunUrl } = require("./workflow_metadata_helpers.cjs");
const { generateHistoryUrl } = require("./generate_history_link.cjs");
const { MAX_LABELS, MAX_ASSIGNEES } = require("./constants.cjs");

const ISSUE_FIELD_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

/**
 * Normalize and validate issue fields payload for update_issue.
 * Ensures fields are objects with a non-empty name and string/number value.
 * @param {any} fields
 * @returns {Array<{name: string, value: string|number}>}
 */
function normalizeIssueFields(fields) {
  if (fields == null) {
    return [];
  }
  if (!Array.isArray(fields)) {
    throw new Error(`${ERR_VALIDATION}: update_issue 'fields' must be an array of objects`);
  }

  return fields.map((field, index) => {
    if (!field || typeof field !== "object" || Array.isArray(field)) {
      throw new Error(`${ERR_VALIDATION}: update_issue 'fields[${index}]' must be an object with 'name' and 'value'`);
    }

    const name = typeof field.name === "string" ? field.name.trim() : "";
    if (!name) {
      throw new Error(`${ERR_VALIDATION}: update_issue 'fields[${index}].name' must be a non-empty string`);
    }

    if (!Object.prototype.hasOwnProperty.call(field, "value")) {
      throw new Error(`${ERR_VALIDATION}: update_issue 'fields[${index}]' is missing required 'value'`);
    }

    const value = field.value;
    if ((typeof value !== "string" && typeof value !== "number") || (typeof value === "number" && !Number.isFinite(value))) {
      throw new Error(`${ERR_VALIDATION}: update_issue 'fields[${index}].value' for "${name}" must be a string or number`);
    }

    return { name, value };
  });
}

/**
 * Resolve issue node ID from issue number.
 * @param {Object} githubClient
 * @param {string} owner
 * @param {string} repo
 * @param {number} issueNumber
 * @returns {Promise<string>}
 */
async function resolveIssueNodeId(githubClient, owner, repo, issueNumber) {
  const result = await githubClient.graphql(
    `query($owner: String!, $repo: String!, $issueNumber: Int!) {
      repository(owner: $owner, name: $repo) {
        issue(number: $issueNumber) {
          id
        }
      }
    }`,
    { owner, repo, issueNumber }
  );

  const issueId = result?.repository?.issue?.id;
  if (!issueId) {
    throw new Error(`${ERR_VALIDATION}: could not resolve node ID for issue #${issueNumber}`);
  }
  return issueId;
}

/**
 * Fetch issue field metadata from repository.
 * @param {Object} githubClient
 * @param {string} owner
 * @param {string} repo
 * @returns {Promise<Array<any>>}
 */
async function fetchIssueFields(githubClient, owner, repo) {
  const result = await githubClient.graphql(
    `query($owner: String!, $repo: String!) {
      repository(owner: $owner, name: $repo) {
        issueFields(first: 100) {
          nodes {
            __typename
            ... on IssueField {
              id
              name
              dataType
            }
            ... on IssueFieldSingleSelect {
              id
              name
              dataType
              options {
                id
                name
              }
            }
            ... on IssueFieldIteration {
              id
              name
              dataType
              configuration {
                iterations {
                  id
                  title
                }
              }
            }
          }
        }
      }
    }`,
    { owner, repo }
  );

  return Array.isArray(result?.repository?.issueFields?.nodes) ? result.repository.issueFields.nodes.filter(Boolean) : [];
}

/**
 * Build GraphQL setIssueFieldValue mutation input from named field values.
 * @param {Array<{name: string, value: string|number}>} requestedFields
 * @param {Array<any>} availableFields
 * @returns {Array<any>}
 */
function buildIssueFieldMutationInput(requestedFields, availableFields) {
  const availableNames = availableFields.map(field => field?.name).filter(Boolean);

  return requestedFields.map(field => {
    const matchedField = availableFields.find(available => typeof available?.name === "string" && available.name.toLowerCase() === field.name.toLowerCase());
    if (!matchedField) {
      throw new Error(`${ERR_VALIDATION}: unknown issue field "${field.name}". Available fields: ${availableNames.join(", ") || "(none)"}`);
    }

    const dataType = typeof matchedField.dataType === "string" ? matchedField.dataType.toUpperCase() : "TEXT";

    if (dataType === "NUMBER") {
      const numberValue = Number(field.value);
      if (!Number.isFinite(numberValue)) {
        throw new Error(`${ERR_VALIDATION}: issue field "${field.name}" requires a numeric value`);
      }
      return { fieldId: matchedField.id, numberValue };
    }

    if (dataType === "DATE") {
      if (typeof field.value !== "string" || !ISSUE_FIELD_DATE_PATTERN.test(field.value)) {
        throw new Error(`${ERR_VALIDATION}: issue field "${field.name}" requires a date value in YYYY-MM-DD format`);
      }
      return { fieldId: matchedField.id, dateValue: field.value };
    }

    if (dataType === "SINGLE_SELECT") {
      const options = Array.isArray(matchedField.options) ? matchedField.options : [];
      const selectedOption = options.find(option => typeof option?.name === "string" && option.name.toLowerCase() === String(field.value).toLowerCase());
      if (!selectedOption) {
        throw new Error(`${ERR_VALIDATION}: invalid option "${field.value}" for issue field "${field.name}". Available options: ${options.map(option => option.name).join(", ") || "(none)"}`);
      }
      return { fieldId: matchedField.id, singleSelectOptionId: selectedOption.id };
    }

    if (dataType === "ITERATION") {
      const iterations = matchedField?.configuration?.iterations;
      const availableIterations = Array.isArray(iterations) ? iterations : [];
      const selectedIteration = availableIterations.find(iteration => typeof iteration?.title === "string" && iteration.title.toLowerCase() === String(field.value).toLowerCase());
      if (!selectedIteration) {
        throw new Error(`${ERR_VALIDATION}: invalid iteration "${field.value}" for issue field "${field.name}". Available iterations: ${availableIterations.map(iteration => iteration.title).join(", ") || "(none)"}`);
      }
      return { fieldId: matchedField.id, singleSelectOptionId: selectedIteration.id };
    }

    return { fieldId: matchedField.id, textValue: String(field.value) };
  });
}

/**
 * Apply issue field values to an existing issue.
 * @param {{githubClient: Object, owner: string, repo: string, issueNumber: number, fields: Array<{name: string, value: string|number}>}} params
 * @returns {Promise<void>}
 */
async function applyIssueFields({ githubClient, owner, repo, issueNumber, fields }) {
  if (!Array.isArray(fields) || fields.length === 0) {
    return;
  }

  if (typeof githubClient.graphql !== "function") {
    throw new Error(`${ERR_VALIDATION}: update_issue 'fields' requires GraphQL access`);
  }

  const issueId = await resolveIssueNodeId(githubClient, owner, repo, issueNumber);
  const availableFields = await fetchIssueFields(githubClient, owner, repo);
  const issueFields = buildIssueFieldMutationInput(fields, availableFields);

  await githubClient.graphql(
    `mutation($input: SetIssueFieldValueInput!) {
      setIssueFieldValue(input: $input) {
        issue {
          id
        }
      }
    }`,
    {
      input: {
        issueId,
        issueFields,
      },
    }
  );
}

/**
 * Execute the issue update API call
 * @param {any} github - GitHub API client
 * @param {any} context - GitHub Actions context
 * @param {number} issueNumber - Issue number to update
 * @param {any} updateData - Data to update
 * @returns {Promise<any>} Updated issue
 */
async function executeIssueUpdate(github, context, issueNumber, updateData) {
  // Handle body operation (append/prepend/replace/replace-island)
  // Default to "append" to add footer with AI attribution
  const operation = updateData._operation || "append";
  let rawBody = updateData._rawBody;
  const includeFooter = updateData._includeFooter !== false; // Default to true
  const titlePrefix = updateData._titlePrefix || "";

  // Remove internal fields
  const { _operation, _rawBody, _includeFooter, _titlePrefix, _workflowRepo, fields, ...apiData } = updateData;

  // Fetch current issue if needed (title prefix validation or body update)
  if (titlePrefix || rawBody !== undefined) {
    const { data: currentIssue } = await github.rest.issues.get({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
    });

    // Validate title prefix if specified
    if (titlePrefix) {
      const currentTitle = currentIssue.title || "";
      if (!currentTitle.startsWith(titlePrefix)) {
        throw new Error(`${ERR_VALIDATION}: Issue title "${currentTitle}" does not start with required prefix "${titlePrefix}"`);
      }
      core.info(`✓ Title prefix validation passed: "${titlePrefix}"`);
    }

    if (rawBody !== undefined) {
      // Load and apply temporary project URL replacements FIRST
      // This resolves any temporary project IDs (e.g., #aw_abc123def456) to actual project URLs
      const temporaryProjectMap = loadTemporaryProjectMap();
      if (temporaryProjectMap.size > 0) {
        rawBody = replaceTemporaryProjectReferences(rawBody, temporaryProjectMap);
        core.debug(`Applied ${temporaryProjectMap.size} temporary project URL replacement(s)`);
      }

      const currentBody = currentIssue.body || "";

      // Get workflow run URL for AI attribution.
      // Use the original workflow repo (_workflowRepo) rather than context.repo, because
      // context may be effectiveContext with repo overridden to a cross-repo target.
      const workflowName = process.env.GH_AW_WORKFLOW_NAME || "GitHub Agentic Workflow";
      const workflowId = process.env.GH_AW_WORKFLOW_ID || "";
      const callerWorkflowId = process.env.GH_AW_CALLER_WORKFLOW_ID || "";
      const workflowRepo = _workflowRepo || context.repo;
      const runUrl = buildWorkflowRunUrl(context, workflowRepo);

      const historyUrl =
        generateHistoryUrl({
          owner: context.repo.owner,
          repo: context.repo.repo,
          itemType: "issue",
          workflowCallId: callerWorkflowId,
          workflowId,
          serverUrl: context.serverUrl,
        }) || undefined;

      // Use helper to update body (handles all operations including replace)
      apiData.body = updateBody({
        currentBody,
        newContent: rawBody,
        operation,
        workflowName,
        runUrl,
        workflowId,
        includeFooter, // Pass footer flag to helper
        historyUrl,
      });

      core.info(`Will update body (length: ${apiData.body.length})`);
    }
  }

  let issue;
  if (Object.keys(apiData).length > 0) {
    const { data } = await github.rest.issues.update({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      ...apiData,
    });
    issue = data;
  } else {
    const { data } = await github.rest.issues.get({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
    });
    issue = data;
  }

  if (Array.isArray(fields) && fields.length > 0) {
    await applyIssueFields({
      githubClient: github,
      owner: context.repo.owner,
      repo: context.repo.repo,
      issueNumber,
      fields,
    });
    core.info(`Applied ${fields.length} issue field(s) to issue #${issueNumber}`);
  }

  return issue;
}

/**
 * Resolve issue number from message and configuration
 * Uses the standard resolve helper for consistency with update_pull_request
 */
const resolveIssueNumber = createStandardResolveNumber({
  itemType: "update_issue",
  itemNumberField: "issue_number",
  supportsPR: false, // Not used when supportsIssue is true
  supportsIssue: true, // update_issue only supports issues, not PRs
});

/**
 * Build update data from message
 * @param {Object} item - The message item
 * @param {Object} config - Configuration object
 * @returns {{success: true, data: Object} | {success: false, error: string}} Update data result
 */
function buildIssueUpdateData(item, config) {
  const updateData = {};

  if (item.title !== undefined) {
    // Sanitize title for Unicode security
    updateData.title = sanitizeTitle(item.title);
  }
  // Check if body updates are allowed (defaults to true if not specified)
  const canUpdateBody = config.allow_body !== false;
  if (item.body !== undefined && canUpdateBody) {
    // Store operation information for consistent footer/append behavior.
    // Default to "append" so we preserve the original issue text.
    updateData._operation = item.operation || "append";
    updateData._rawBody = item.body;
  } else if (item.body !== undefined && !canUpdateBody) {
    // Body update attempted but not allowed by configuration
    core.warning("Body update not allowed by safe-outputs configuration");
  }
  // The safe-outputs schema uses "status" (open/closed), while the GitHub API uses "state".
  // Accept both for compatibility.
  if (item.state !== undefined) {
    updateData.state = item.state;
  } else if (item.status !== undefined) {
    updateData.state = item.status;
  }
  if (item.labels !== undefined) {
    updateData.labels = item.labels;
  }
  if (item.assignees !== undefined) {
    updateData.assignees = item.assignees;
  }
  if (item.milestone !== undefined) {
    updateData.milestone = item.milestone;
  }
  if (item.fields !== undefined) {
    try {
      updateData.fields = normalizeIssueFields(item.fields);
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : String(error) };
    }
  }

  // Enforce max limits on labels and assignees before API calls
  const labelsLimitResult = tryEnforceArrayLimit(updateData.labels, MAX_LABELS, "labels");
  if (!labelsLimitResult.success) {
    core.warning(`Issue update limit exceeded: ${labelsLimitResult.error}`);
    return { success: false, error: labelsLimitResult.error };
  }

  const assigneesLimitResult = tryEnforceArrayLimit(updateData.assignees, MAX_ASSIGNEES, "assignees");
  if (!assigneesLimitResult.success) {
    core.warning(`Issue update limit exceeded: ${assigneesLimitResult.error}`);
    return { success: false, error: assigneesLimitResult.error };
  }

  // Pass footer config to executeUpdate (default to true)
  updateData._includeFooter = parseBoolTemplatable(config.footer, true);

  // Store title prefix for validation in executeIssueUpdate
  if (config.title_prefix) {
    updateData._titlePrefix = config.title_prefix;
  }

  return { success: true, data: updateData };
}

/**
 * Format success result for issue update
 * Uses the standard format helper for consistency across update handlers
 */
const formatIssueSuccessResult = createStandardFormatResult({
  numberField: "number",
  urlField: "url",
  urlSource: "html_url",
});

/**
 * Main handler factory for update_issue
 * Returns a message handler function that processes individual update_issue messages
 * @type {HandlerFactoryFunction}
 */
const main = createUpdateHandlerFactory({
  itemType: "update_issue",
  itemTypeName: "issue",
  supportsPR: false, // Not used by factory, but kept for documentation
  resolveItemNumber: resolveIssueNumber,
  buildUpdateData: buildIssueUpdateData,
  executeUpdate: executeIssueUpdate,
  formatSuccessResult: formatIssueSuccessResult,
});

module.exports = { main, buildIssueUpdateData };
