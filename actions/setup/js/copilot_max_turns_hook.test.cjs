import { afterEach, describe, expect, it } from "vitest";
import fs from "fs";
import os from "os";
import path from "path";
import { createRequire } from "module";

const require = createRequire(import.meta.url);
const { applyMaxTurnsGuardrail, detectHookEvent, getSessionID, parseMaxTurns, readTurnCount, sanitizeSessionID, writeTurnCount } = require("./copilot_max_turns_hook.cjs");

describe("copilot_max_turns_hook.cjs", () => {
  describe("parseMaxTurns", () => {
    it("returns parsed positive integer", () => {
      expect(parseMaxTurns("3")).toBe(3);
      expect(parseMaxTurns(" 10 ")).toBe(10);
    });

    it("returns null for missing, zero, and invalid values", () => {
      expect(parseMaxTurns(undefined)).toBeNull();
      expect(parseMaxTurns("0")).toBeNull();
      expect(parseMaxTurns("-1")).toBeNull();
      expect(parseMaxTurns("abc")).toBeNull();
    });
  });

  describe("event detection", () => {
    it("detects sessionStart payload", () => {
      expect(detectHookEvent({ sessionId: "s1", source: "new" })).toBe("sessionStart");
    });

    it("detects agentStop payload", () => {
      expect(detectHookEvent({ sessionId: "s1", stopReason: "end_turn" })).toBe("agentStop");
    });

    it("detects preToolUse payload", () => {
      expect(detectHookEvent({ sessionId: "s1", toolName: "bash" })).toBe("preToolUse");
    });
  });

  describe("applyMaxTurnsGuardrail", () => {
    it("resets count on session start", () => {
      const result = applyMaxTurnsGuardrail({ source: "new" }, 5, 3);
      expect(result).toEqual({ nextTurnCount: 0 });
    });

    it("increments count on agent stop", () => {
      const result = applyMaxTurnsGuardrail({ stopReason: "end_turn" }, 5, 2);
      expect(result).toEqual({ nextTurnCount: 3 });
    });

    it("denies tool use once turn limit is reached", () => {
      const result = applyMaxTurnsGuardrail({ toolName: "bash" }, 2, 2);
      expect(result.nextTurnCount).toBe(2);
      expect(result.denyReason).toContain("Reached maximum number of turns (2)");
    });
  });

  describe("state helpers", () => {
    const tempDir = path.join(os.tmpdir(), `copilot-max-turns-hook-${Date.now()}`);
    const stateFile = path.join(tempDir, `${sanitizeSessionID("session/one")}.json`);

    afterEach(() => {
      fs.rmSync(tempDir, { recursive: true, force: true });
    });

    it("reads and writes turn counts", () => {
      expect(readTurnCount(stateFile)).toBe(0);
      writeTurnCount(stateFile, 4);
      expect(readTurnCount(stateFile)).toBe(4);
    });
  });

  describe("session id extraction", () => {
    it("supports camelCase and snake_case ids", () => {
      expect(getSessionID({ sessionId: "abc" })).toBe("abc");
      expect(getSessionID({ session_id: "def" })).toBe("def");
    });

    it("falls back to default when missing", () => {
      expect(getSessionID({})).toBe("default");
      expect(getSessionID(null)).toBe("default");
    });
  });
});
