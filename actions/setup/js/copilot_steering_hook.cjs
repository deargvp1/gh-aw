// @ts-check

"use strict";

const fs = require("fs");

const DEFAULT_TIMEOUT_MINUTES = 30;
const DEFAULT_TIME_WARNING_MINUTES = 5;
const DEFAULT_TIME_CRITICAL_MINUTES = 2;
const DEFAULT_RUN_WARNING_REMAINING = 2;
const DEFAULT_RUN_CRITICAL_REMAINING = 1;
const DEFAULT_STATE_PATH = "/tmp/gh-aw/copilot-steering-state.json";

/**
 * @typedef {{
 *   startedAtMs: number,
 *   turns: number,
 *   warningInjected: boolean,
 *   criticalInjected: boolean
 * }} SteeringState
 */

/**
 * @param {string | undefined} rawValue
 * @param {number} fallback
 * @returns {number}
 */
function parsePositiveNumber(rawValue, fallback) {
  const parsed = parseFloat(rawValue || "");
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

/**
 * @param {NodeJS.ProcessEnv} env
 * @returns {{
 *   timeoutMinutes: number,
 *   timeWarningMinutes: number,
 *   timeCriticalMinutes: number,
 *   runsWarningRemaining: number,
 *   runsCriticalRemaining: number,
 *   maxRuns: number,
 *   statePath: string
 * }}
 */
function loadSteeringConfig(env = process.env) {
  return {
    timeoutMinutes: parsePositiveNumber(env.GH_AW_TIMEOUT_MINUTES, DEFAULT_TIMEOUT_MINUTES),
    timeWarningMinutes: parsePositiveNumber(env.GH_AW_STEERING_TIME_WARNING_MINUTES, DEFAULT_TIME_WARNING_MINUTES),
    timeCriticalMinutes: parsePositiveNumber(env.GH_AW_STEERING_TIME_CRITICAL_MINUTES, DEFAULT_TIME_CRITICAL_MINUTES),
    runsWarningRemaining: parsePositiveNumber(env.GH_AW_STEERING_RUN_WARNING_REMAINING, DEFAULT_RUN_WARNING_REMAINING),
    runsCriticalRemaining: parsePositiveNumber(env.GH_AW_STEERING_RUN_CRITICAL_REMAINING, DEFAULT_RUN_CRITICAL_REMAINING),
    maxRuns: parsePositiveNumber(env.GH_AW_COPILOT_MAX_RUNS, 0),
    statePath: env.GH_AW_COPILOT_STEERING_STATE_PATH || DEFAULT_STATE_PATH,
  };
}

/**
 * @param {number} timestamp
 * @returns {SteeringState}
 */
function createInitialState(timestamp) {
  return {
    startedAtMs: timestamp,
    turns: 0,
    warningInjected: false,
    criticalInjected: false,
  };
}

/**
 * @param {string} statePath
 * @returns {SteeringState | null}
 */
function loadState(statePath) {
  try {
    if (!fs.existsSync(statePath)) {
      return null;
    }
    const raw = fs.readFileSync(statePath, "utf8");
    return /** @type {SteeringState} */ JSON.parse(raw);
  } catch {
    return null;
  }
}

/**
 * @param {string} statePath
 * @param {SteeringState} state
 */
function saveState(statePath, state) {
  fs.writeFileSync(statePath, JSON.stringify(state), "utf8");
}

/**
 * @param {unknown} value
 * @returns {number}
 */
function parseEventTimestamp(value) {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Date.parse(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return Date.now();
}

/**
 * @param {number | null} remainingMinutes
 * @param {number | null} remainingRuns
 * @returns {string}
 */
function buildBudgetSummary(remainingMinutes, remainingRuns) {
  /** @type {string[]} */
  const parts = [];
  if (remainingMinutes !== null) {
    parts.push(`${Math.max(0, remainingMinutes).toFixed(1)} minute(s) left`);
  }
  if (remainingRuns !== null) {
    parts.push(`${Math.max(0, remainingRuns)} run(s) left`);
  }
  return parts.join(", ");
}

/**
 * @param {SteeringState} state
 * @param {{
 *   timeoutMinutes: number,
 *   timeWarningMinutes: number,
 *   timeCriticalMinutes: number,
 *   runsWarningRemaining: number,
 *   runsCriticalRemaining: number,
 *   maxRuns: number
 * }} config
 * @param {number} timestamp
 * @returns {{ state: SteeringState, decision: { decision: "block", reason: string } | null }}
 */
function computeSteeringDecision(state, config, timestamp) {
  const nextState = { ...state, turns: state.turns + 1 };
  const elapsedMinutes = (timestamp - nextState.startedAtMs) / 60000;
  const remainingMinutes = Number.isFinite(config.timeoutMinutes) ? config.timeoutMinutes - elapsedMinutes : null;
  const remainingRuns = config.maxRuns > 0 ? config.maxRuns - nextState.turns : null;

  const isCriticalTime = remainingMinutes !== null && remainingMinutes <= config.timeCriticalMinutes;
  const isWarningTime = remainingMinutes !== null && remainingMinutes <= config.timeWarningMinutes;
  const isCriticalRuns = remainingRuns !== null && remainingRuns <= config.runsCriticalRemaining;
  const isWarningRuns = remainingRuns !== null && remainingRuns <= config.runsWarningRemaining;
  const budgetSummary = buildBudgetSummary(remainingMinutes, remainingRuns);

  if (!nextState.criticalInjected && (isCriticalTime || isCriticalRuns)) {
    nextState.warningInjected = true;
    nextState.criticalInjected = true;
    return {
      state: nextState,
      decision: {
        decision: "block",
        reason: `⚠️ CRITICAL: Budget is nearly exhausted (${budgetSummary}). Stop new exploration and produce your final output now.`,
      },
    };
  }

  if (!nextState.warningInjected && (isWarningTime || isWarningRuns)) {
    nextState.warningInjected = true;
    return {
      state: nextState,
      decision: {
        decision: "block",
        reason: `⚠️ Warning: Budget is getting low (${budgetSummary}). Wrap up your work and move to final output.`,
      },
    };
  }

  return { state: nextState, decision: null };
}

/**
 * @param {"sessionStart" | "agentStop"} eventName
 * @param {Record<string, any>} payload
 * @param {NodeJS.ProcessEnv} env
 * @returns {{ state: SteeringState, decision: { decision: "block", reason: string } | null }}
 */
function handleSteeringEvent(eventName, payload, env = process.env) {
  const config = loadSteeringConfig(env);
  const timestamp = parseEventTimestamp(payload.timestamp);
  const priorState = loadState(config.statePath) || createInitialState(timestamp);

  if (eventName === "sessionStart") {
    const isNewSession = payload.source === "new";
    const state = isNewSession ? createInitialState(timestamp) : priorState;
    saveState(config.statePath, state);
    return { state, decision: null };
  }

  const result = computeSteeringDecision(priorState, config, timestamp);
  saveState(config.statePath, result.state);
  return result;
}

/**
 * @returns {Record<string, any>}
 */
function readStdinJSON() {
  try {
    const input = fs.readFileSync(0, "utf8").trim();
    return input ? JSON.parse(input) : {};
  } catch {
    return {};
  }
}

function main() {
  const eventName = process.argv[2];
  if (eventName !== "sessionStart" && eventName !== "agentStop") {
    process.stderr.write(`[copilot-steering-hook] unsupported event: ${String(eventName || "")}\n`);
    return;
  }

  const payload = readStdinJSON();
  const { decision } = handleSteeringEvent(eventName, payload, process.env);
  if (decision) {
    process.stdout.write(JSON.stringify(decision));
  }
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = {
    computeSteeringDecision,
    createInitialState,
    handleSteeringEvent,
    loadSteeringConfig,
    parseEventTimestamp,
  };
}

if (require.main === module) {
  main();
}
