// @ts-check
/// <reference types="@actions/github-script" />

const fs = require("fs");
const { main: exportCopilotOtelTraces } = require("./export_copilot_otel_traces.cjs");

const AW_INFO_PATH = "/tmp/gh-aw/aw_info.json";
const AGENT_OUTPUT_PATH = "/tmp/gh-aw/agent_output.json";
const OTEL_JSONL_PATH = "/tmp/gh-aw/otel.jsonl";
const gatewayEventPaths = ["/tmp/gh-aw/mcp-logs/gateway.jsonl", "/tmp/gh-aw/mcp-logs/rpc-messages.jsonl"];

function getAttributeValue(attributes, key) {
  if (!Array.isArray(attributes)) {
    return "";
  }

  const attribute = attributes.find(attr => attr && attr.key === key && attr.value);
  if (!attribute || !attribute.value || typeof attribute.value !== "object") {
    return "";
  }

  return attribute.value.stringValue || attribute.value.intValue || attribute.value.doubleValue || "";
}

function formatSpanKind(kind) {
  switch (kind) {
    case 2:
      return "SERVER";
    case 3:
      return "CLIENT";
    case 4:
      return "PRODUCER";
    case 5:
      return "CONSUMER";
    default:
      return "INTERNAL";
  }
}

function formatSpanStatus(status) {
  if (!status || typeof status !== "object") {
    return "";
  }
  if (status.code === 2) {
    return "ERROR";
  }
  if (status.code === 1) {
    return "OK";
  }
  return "UNSET";
}

function collectTracePreview(traceId = "") {
  if (!fs.existsSync(OTEL_JSONL_PATH)) {
    return null;
  }

  const preview = [];
  let totalSpans = 0;
  const lines = fs.readFileSync(OTEL_JSONL_PATH, "utf8").split("\n");
  for (const raw of lines) {
    const line = raw.trim();
    if (!line) continue;

    let payload;
    try {
      payload = JSON.parse(line);
    } catch {
      continue;
    }

    const resourceSpans = Array.isArray(payload?.resourceSpans) ? payload.resourceSpans : [];
    for (const resourceSpan of resourceSpans) {
      const serviceName = getAttributeValue(resourceSpan?.resource?.attributes, "service.name");
      const scopeSpans = Array.isArray(resourceSpan?.scopeSpans) ? resourceSpan.scopeSpans : [];
      for (const scopeSpan of scopeSpans) {
        const spans = Array.isArray(scopeSpan?.spans) ? scopeSpan.spans : [];
        for (const span of spans) {
          if (traceId && span?.traceId && span.traceId !== traceId) {
            continue;
          }
          totalSpans++;
          if (preview.length < 8) {
            preview.push({
              name: span?.name || "unknown",
              kind: formatSpanKind(span?.kind),
              status: formatSpanStatus(span?.status),
              serviceName: serviceName || "unknown",
            });
          }
        }
      }
    }
  }

  if (totalSpans === 0 && preview.length === 0) {
    return null;
  }

  return { totalSpans, spans: preview };
}

function readJSONIfExists(path) {
  if (!fs.existsSync(path)) {
    return null;
  }

  try {
    return JSON.parse(fs.readFileSync(path, "utf8"));
  } catch {
    return null;
  }
}

function countBlockedRequests() {
  let total = 0;

  for (const path of gatewayEventPaths) {
    if (!fs.existsSync(path)) {
      continue;
    }

    const lines = fs.readFileSync(path, "utf8").split("\n");
    for (const raw of lines) {
      const line = raw.trim();
      if (!line) continue;
      try {
        const entry = JSON.parse(line);
        if (entry && entry.type === "DIFC_FILTERED") total++;
      } catch {
        // skip malformed lines
      }
    }
  }

  return total;
}

function uniqueCreatedItemTypes(items) {
  const types = new Set();

  for (const item of items) {
    if (item && typeof item.type === "string" && item.type.trim() !== "") {
      types.add(item.type);
    }
  }

  return [...types].sort();
}

function collectObservabilityData() {
  const awInfo = readJSONIfExists(AW_INFO_PATH) || {};
  const agentOutput = readJSONIfExists(AGENT_OUTPUT_PATH) || { items: [], errors: [] };
  const items = Array.isArray(agentOutput.items) ? agentOutput.items : [];
  const errors = Array.isArray(agentOutput.errors) ? agentOutput.errors : [];
  // Prefer GITHUB_AW_OTEL_TRACE_ID (written to GITHUB_ENV by action_setup_otlp.cjs)
  // so the summary always shows the trace ID that is actually present in the OTLP backend.
  // Fall back to context.otel_trace_id for cross-workflow traces propagated from a parent.
  // Do NOT fall back to workflow_call_id — it is not a valid OTLP trace ID.
  const traceId = process.env.GITHUB_AW_OTEL_TRACE_ID || (awInfo.context ? awInfo.context.otel_trace_id || "" : "");

  return {
    workflowName: awInfo.workflow_name || "",
    engineId: awInfo.engine_id || "",
    traceId,
    tracePreview: collectTracePreview(traceId),
    staged: awInfo.staged === true,
    firewallEnabled: awInfo.firewall_enabled === true,
    createdItemCount: items.length,
    createdItemTypes: uniqueCreatedItemTypes(items),
    outputErrorCount: errors.length,
    blockedRequests: countBlockedRequests(),
  };
}

function buildObservabilitySummary(data) {
  const posture = data.createdItemCount > 0 ? "write-capable" : "read-only";
  const lines = [];

  lines.push("<details>");
  lines.push("<summary>Observability</summary>");
  lines.push("");

  if (data.workflowName) {
    lines.push(`- **workflow**: ${data.workflowName}`);
  }
  if (data.engineId) {
    lines.push(`- **engine**: ${data.engineId}`);
  }
  if (data.traceId) {
    lines.push(`- **trace id**: ${data.traceId}`);
  }

  lines.push(`- **posture**: ${posture}`);
  lines.push(`- **created items**: ${data.createdItemCount}`);
  lines.push(`- **blocked requests**: ${data.blockedRequests}`);
  lines.push(`- **agent output errors**: ${data.outputErrorCount}`);
  lines.push(`- **firewall enabled**: ${data.firewallEnabled}`);
  lines.push(`- **staged**: ${data.staged}`);

  if (data.createdItemTypes.length > 0) {
    lines.push("- **item types**:");
    for (const itemType of data.createdItemTypes) {
      lines.push(`  - ${itemType}`);
    }
  }

  if (data.tracePreview && data.tracePreview.totalSpans > 0) {
    lines.push("");
    lines.push("<details>");
    lines.push("<summary>Trace preview</summary>");
    lines.push("");
    lines.push(`**Spans captured:** ${data.tracePreview.totalSpans}`);
    lines.push("");
    lines.push("| Span | Kind | Status | Service |");
    lines.push("|------|------|--------|---------|");
    for (const span of data.tracePreview.spans) {
      lines.push(`| ${span.name} | ${span.kind} | ${span.status || ""} | ${span.serviceName} |`);
    }
    lines.push("</details>");
  }

  lines.push("");
  lines.push("</details>");

  return lines.join("\n") + "\n";
}

async function main(core) {
  await exportCopilotOtelTraces(core);
  const data = collectObservabilityData();
  const markdown = buildObservabilitySummary(data);
  await core.summary.addRaw(markdown).write();
  core.info("Generated observability summary in step summary");
}

module.exports = {
  buildObservabilitySummary,
  collectTracePreview,
  collectObservabilityData,
  main,
};
