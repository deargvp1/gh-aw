// @ts-check
"use strict";

const fs = require("fs");
const path = require("path");

const DEFAULT_STATE_DIR = "/tmp/gh-aw/copilot-hooks/max-turns";
const STATE_VERSION = 1;

/**
 * @param {string | undefined} value
 * @returns {number | null}
 */
function parseMaxTurns(value) {
  if (!value) return null;
  const parsed = Number.parseInt(value.trim(), 10);
  if (!Number.isFinite(parsed) || parsed <= 0) return null;
  return parsed;
}

/**
 * @param {any} payload
 * @returns {string}
 */
function getSessionID(payload) {
  const sessionID = payload && (payload.sessionId || payload.session_id);
  if (typeof sessionID !== "string" || sessionID.trim() === "") {
    return "default";
  }
  return sessionID.trim();
}

/**
 * @param {any} payload
 * @returns {"sessionStart" | "agentStop" | "preToolUse" | "other"}
 */
function detectHookEvent(payload) {
  if (!payload || typeof payload !== "object") {
    return "other";
  }
  if ("source" in payload && !("toolName" in payload || "tool_name" in payload)) {
    return "sessionStart";
  }
  if ("stopReason" in payload || "stop_reason" in payload) {
    return "agentStop";
  }
  if ("toolName" in payload || "tool_name" in payload) {
    return "preToolUse";
  }
  return "other";
}

/**
 * @param {string} sessionID
 * @returns {string}
 */
function sanitizeSessionID(sessionID) {
  return sessionID.replace(/[^a-zA-Z0-9._-]/g, "_");
}

/**
 * @param {string} stateDir
 * @param {string} sessionID
 * @returns {string}
 */
function getStateFilePath(stateDir, sessionID) {
  return path.join(stateDir, `${sanitizeSessionID(sessionID)}.json`);
}

/**
 * @param {string} stateFile
 * @returns {number}
 */
function readTurnCount(stateFile) {
  try {
    const content = fs.readFileSync(stateFile, "utf8");
    const parsed = JSON.parse(content);
    if (!parsed || typeof parsed.turnCount !== "number" || parsed.turnCount < 0) {
      return 0;
    }
    return Math.floor(parsed.turnCount);
  } catch {
    return 0;
  }
}

/**
 * @param {string} stateFile
 * @param {number} turnCount
 * @returns {void}
 */
function writeTurnCount(stateFile, turnCount) {
  fs.mkdirSync(path.dirname(stateFile), { recursive: true });
  const payload = JSON.stringify({
    version: STATE_VERSION,
    turnCount: Math.max(0, Math.floor(turnCount)),
  });
  fs.writeFileSync(stateFile, payload, { encoding: "utf8" });
}

/**
 * @param {any} payload
 * @param {number} maxTurns
 * @param {number} currentTurnCount
 * @returns {{nextTurnCount: number, denyReason?: string}}
 */
function applyMaxTurnsGuardrail(payload, maxTurns, currentTurnCount) {
  const event = detectHookEvent(payload);

  if (event === "sessionStart") {
    return { nextTurnCount: 0 };
  }

  if (event === "agentStop") {
    return { nextTurnCount: currentTurnCount + 1 };
  }

  if (event === "preToolUse" && currentTurnCount >= maxTurns) {
    return {
      nextTurnCount: currentTurnCount,
      denyReason: `Reached maximum number of turns (${maxTurns}). Stopping.`,
    };
  }

  return { nextTurnCount: currentTurnCount };
}

/**
 * @returns {Promise<void>}
 */
async function main() {
  const maxTurns = parseMaxTurns(process.env.GH_AW_MAX_TURNS);
  if (maxTurns === null) {
    return;
  }

  const input = await new Promise(resolve => {
    let raw = "";
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", chunk => {
      raw += chunk;
    });
    process.stdin.on("end", () => resolve(raw));
  });

  if (!input || input.trim() === "") {
    return;
  }

  let payload;
  try {
    payload = JSON.parse(input);
  } catch {
    return;
  }

  const sessionID = getSessionID(payload);
  const stateDir = process.env.GH_AW_COPILOT_MAX_TURNS_STATE_DIR || DEFAULT_STATE_DIR;
  const stateFile = getStateFilePath(stateDir, sessionID);
  const currentTurnCount = readTurnCount(stateFile);
  const result = applyMaxTurnsGuardrail(payload, maxTurns, currentTurnCount);

  writeTurnCount(stateFile, result.nextTurnCount);

  if (result.denyReason) {
    process.stdout.write(
      JSON.stringify({
        permissionDecision: "deny",
        permissionDecisionReason: result.denyReason,
      })
    );
  }
}

module.exports = {
  DEFAULT_STATE_DIR,
  applyMaxTurnsGuardrail,
  detectHookEvent,
  getSessionID,
  parseMaxTurns,
  readTurnCount,
  sanitizeSessionID,
  writeTurnCount,
};

if (require.main === module) {
  main().catch(() => {
    process.exit(0);
  });
}
