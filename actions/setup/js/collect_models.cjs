// @ts-check
/// <reference types="@actions/github-script" />

const fs = require("fs");
const path = require("path");

const REQUEST_TIMEOUT_MS = 15000;
const MAX_MODELS_DISPLAY = 100;

function ensureParent(filePath) {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });
}

function normalizeURL(host, route) {
  const base = host.startsWith("http://") || host.startsWith("https://") ? host : `https://${host}`;
  const url = new URL(base);
  const normalizedRoute = route.startsWith("/") ? route : `/${route}`;
  url.pathname = normalizedRoute;
  url.search = "";
  return url;
}

async function requestJSON(url, headers) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(new Error("request timeout")), REQUEST_TIMEOUT_MS);
  try {
    const response = await fetch(url, {
      method: "GET",
      headers,
      signal: controller.signal,
    });
    const body = await response.text();
    return { statusCode: response.status, body };
  } finally {
    clearTimeout(timeoutId);
  }
}

function extractModels(payload) {
  if (Array.isArray(payload?.data)) {
    return payload.data;
  }
  if (Array.isArray(payload?.models)) {
    return payload.models;
  }
  return [];
}

function modelDisplayName(model) {
  if (typeof model === "string") {
    return model;
  }
  if (model && typeof model === "object") {
    return model.id || model.name || model.model || JSON.stringify(model);
  }
  return String(model);
}

function buildHeaders(authType, token) {
  const headers = {
    Accept: "application/json",
    "User-Agent": "gh-aw-models-collector",
  };

  if (!token) {
    return headers;
  }

  if (authType === "x-api-key") {
    headers["x-api-key"] = token;
    headers["anthropic-version"] = "2023-06-01";
    return headers;
  }

  if (authType === "x-goog-api-key") {
    headers["x-goog-api-key"] = token;
    return headers;
  }

  headers.Authorization = `Bearer ${token}`;
  return headers;
}

async function writeSummary(engineId, models, errorMessage) {
  if (errorMessage) {
    core.summary.addDetails(`Available Models (${engineId})`, `\n\n⚠️ ${errorMessage}\n`);
    await core.summary.write();
    return;
  }

  const names = models.map(modelDisplayName);
  // Keep summaries compact to avoid GitHub step summary size pressure.
  const lines = names
    .slice(0, MAX_MODELS_DISPLAY)
    .map(name => `- \`${name}\``)
    .join("\n");
  const hidden = names.length > MAX_MODELS_DISPLAY ? `\n\n_...and ${names.length - MAX_MODELS_DISPLAY} more._` : "";
  const body = names.length > 0 ? `\n\n${lines}${hidden}\n` : "\n\n_No models were returned by the provider._\n";
  core.summary.addDetails(`Available Models (${engineId}, ${names.length})`, body);
  await core.summary.write();
}

async function main() {
  const engineId = process.env.GH_AW_ENGINE_ID || "unknown";
  const modelsHost = process.env.GH_AW_MODELS_HOST || "";
  const modelsRoute = process.env.GH_AW_MODELS_ROUTE || "";
  const modelsFile = process.env.GH_AW_MODELS_FILE || "";
  const artifactFile = process.env.GH_AW_MODELS_ARTIFACT_FILE || "";
  const authType = process.env.GH_AW_MODELS_AUTH_TYPE || "bearer";
  const token = process.env.GH_AW_MODELS_TOKEN || "";

  if (!modelsHost || !modelsRoute || !modelsFile) {
    core.info("Skipping model collection: host, route, or output path missing");
    return;
  }

  try {
    const url = normalizeURL(modelsHost, modelsRoute);
    const headers = buildHeaders(authType, token);
    const response = await requestJSON(url, headers);

    if (response.statusCode === 401 || response.statusCode === 403) {
      const warning = `Could not fetch models from ${url.host}${url.pathname}: invalid or unauthorized token (HTTP ${response.statusCode})`;
      core.warning(warning);
      await writeSummary(engineId, [], warning);
      return;
    }

    if (response.statusCode < 200 || response.statusCode >= 300) {
      const warning = `Could not fetch models from ${url.host}${url.pathname}: HTTP ${response.statusCode}`;
      core.warning(warning);
      await writeSummary(engineId, [], warning);
      return;
    }

    let payload = {};
    try {
      payload = JSON.parse(response.body);
    } catch (error) {
      const warning = `Could not parse models response JSON: ${error instanceof Error ? error.message : String(error)}`;
      core.warning(warning);
      await writeSummary(engineId, [], warning);
      return;
    }

    const models = extractModels(payload);
    ensureParent(modelsFile);
    fs.writeFileSync(modelsFile, JSON.stringify(payload, null, 2) + "\n", "utf8");
    core.info(`Saved models response to ${modelsFile}`);

    if (artifactFile) {
      ensureParent(artifactFile);
      fs.copyFileSync(modelsFile, artifactFile);
      core.info(`Copied models response to artifact path ${artifactFile}`);
    }

    await writeSummary(engineId, models, undefined);
  } catch (error) {
    const warning = `Model collection failed: ${error instanceof Error ? error.message : String(error)}`;
    core.warning(warning);
    await writeSummary(engineId, [], warning);
  }
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = {
    main,
    normalizeURL,
    extractModels,
    modelDisplayName,
    buildHeaders,
  };
}

if (require.main === module) {
  main();
}
