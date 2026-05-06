import { afterEach, describe, expect, it } from "vitest";
import { createRequire } from "module";
import fs from "fs";
import os from "os";
import path from "path";

const require = createRequire(import.meta.url);
const { appendHookEventLog, createInitialState, handleSteeringEvent, loadSteeringConfig } = require("./copilot_steering_hook.cjs");

describe("copilot_steering_hook.cjs", () => {
  let tempDir = "";
  let statePath = "";

  afterEach(() => {
    if (tempDir) {
      fs.rmSync(tempDir, { recursive: true, force: true });
      tempDir = "";
      statePath = "";
    }
  });

  function makeTestEnv(overrides = {}) {
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "copilot-steering-hook-"));
    statePath = path.join(tempDir, "state.json");
    return {
      GH_AW_TIMEOUT_MINUTES: "30",
      GH_AW_STEERING_TIME_WARNING_MINUTES: "5",
      GH_AW_STEERING_TIME_CRITICAL_MINUTES: "2",
      GH_AW_STEERING_RUN_WARNING_REMAINING: "2",
      GH_AW_STEERING_RUN_CRITICAL_REMAINING: "1",
      GH_AW_COPILOT_MAX_RUNS: "4",
      GH_AW_COPILOT_STEERING_STATE_PATH: statePath,
      ...overrides,
    };
  }

  it("loads steering config from environment with defaults fallback", () => {
    const config = loadSteeringConfig({ GH_AW_COPILOT_STEERING_STATE_PATH: "/tmp/state.json" });
    expect(config.timeoutMinutes).toBe(30);
    expect(config.timeWarningMinutes).toBe(5);
    expect(config.timeCriticalMinutes).toBe(2);
    expect(config.runsWarningRemaining).toBe(2);
    expect(config.runsCriticalRemaining).toBe(1);
    expect(config.hookLogPath).toBe("/tmp/gh-aw/copilot-steering-hook.log");
  });

  it("appends hook event log entries as JSON lines", () => {
    makeTestEnv();
    const hookLogPath = path.join(tempDir, "hook.log");
    appendHookEventLog(hookLogPath, {
      event: "sessionStart",
      timestamp: 1234,
      statePath,
      turns: 0,
      warningInjected: false,
      criticalInjected: false,
      decision: "none",
    });
    const content = fs.readFileSync(hookLogPath, "utf8").trim();
    const parsed = JSON.parse(content);
    expect(parsed.event).toBe("sessionStart");
    expect(parsed.decision).toBe("none");
  });

  it("initializes state on sessionStart without emitting a decision", () => {
    const env = makeTestEnv();
    const result = handleSteeringEvent("sessionStart", { timestamp: 1000, source: "new" }, env);
    expect(result.decision).toBeNull();
    expect(result.state).toEqual(createInitialState(1000));
    expect(fs.existsSync(statePath)).toBe(true);
  });

  it("emits warning steering when remaining run budget hits warning threshold", () => {
    const env = makeTestEnv({
      GH_AW_STEERING_TIME_WARNING_MINUTES: "0.1",
      GH_AW_STEERING_TIME_CRITICAL_MINUTES: "0.05",
    });
    handleSteeringEvent("sessionStart", { timestamp: 1000, source: "new" }, env);
    const firstStop = handleSteeringEvent("agentStop", { timestamp: 1100 }, env);
    expect(firstStop.decision).toBeNull();
    const secondStop = handleSteeringEvent("agentStop", { timestamp: 1200 }, env);
    expect(secondStop.decision).not.toBeNull();
    expect(secondStop.decision.decision).toBe("block");
    expect(secondStop.decision.reason).toContain("Warning");
    expect(secondStop.decision.reason).toContain("run(s) left");
  });

  it("emits critical steering when remaining time is below critical threshold", () => {
    const env = makeTestEnv({
      GH_AW_TIMEOUT_MINUTES: "1",
      GH_AW_STEERING_TIME_WARNING_MINUTES: "0.5",
      GH_AW_STEERING_TIME_CRITICAL_MINUTES: "0.2",
      GH_AW_COPILOT_MAX_RUNS: "0",
    });
    handleSteeringEvent("sessionStart", { timestamp: 0, source: "new" }, env);
    const result = handleSteeringEvent("agentStop", { timestamp: 50 * 1000 }, env);
    expect(result.decision).not.toBeNull();
    expect(result.decision.decision).toBe("block");
    expect(result.decision.reason).toContain("CRITICAL");
    expect(result.decision.reason).toContain("minute(s) left");
  });

  it("falls back to initial state when persisted state is malformed-but-parseable", () => {
    const env = makeTestEnv({
      GH_AW_TIMEOUT_MINUTES: "10",
      GH_AW_STEERING_TIME_WARNING_MINUTES: "9.9",
      GH_AW_STEERING_TIME_CRITICAL_MINUTES: "0.1",
      GH_AW_COPILOT_MAX_RUNS: "2",
    });
    fs.writeFileSync(statePath, "{}", "utf8");
    const result = handleSteeringEvent("agentStop", { timestamp: 60 * 1000 }, env);
    expect(result.state.turns).toBe(1);
    expect(result.state.warningInjected).toBe(true);
    expect(result.decision).not.toBeNull();
    expect(result.decision.reason).toContain("minute(s) left");
  });

  it("creates parent directory before saving state", () => {
    const env = makeTestEnv();
    const nestedStatePath = path.join(tempDir, "nested", "deeper", "state.json");
    env.GH_AW_COPILOT_STEERING_STATE_PATH = nestedStatePath;

    const result = handleSteeringEvent("sessionStart", { timestamp: 1000, source: "new" }, env);
    expect(result.decision).toBeNull();
    expect(fs.existsSync(nestedStatePath)).toBe(true);
  });

  it("removes persisted state on sessionEnd", () => {
    const env = makeTestEnv();
    handleSteeringEvent("sessionStart", { timestamp: 1000, source: "new" }, env);
    expect(fs.existsSync(statePath)).toBe(true);

    const result = handleSteeringEvent("sessionEnd", { timestamp: 2000 }, env);
    expect(result.decision).toBeNull();
    expect(fs.existsSync(statePath)).toBe(false);
  });
});
