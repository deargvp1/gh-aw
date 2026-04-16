// @ts-check
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { createRequire } from "module";
import { mkdtempSync, readFileSync, writeFileSync, unlinkSync, existsSync } from "fs";
import { tmpdir } from "os";
import { join } from "path";

const req = createRequire(import.meta.url);

// Load send_otlp_span module and capture originals for restore
const sendOtlpModule = req("./send_otlp_span.cjs");
const originalSendJobSetupSpan = sendOtlpModule.sendJobSetupSpan;

// Load the module under test — it holds a reference to the same sendOtlpModule object
const { run } = req("./action_setup_otlp.cjs");

const VALID_TRACE_ID = "abcdef1234567890abcdef1234567890"; // 32 hex chars
const VALID_SPAN_ID = "abcdef1234567890"; // 16 hex chars
const INVALID_TRACE_ID = "not-a-valid-trace-id";
const INVALID_SPAN_ID = "";

const mockSendJobSetupSpan = vi.fn();

describe("action_setup_otlp.cjs", () => {
  /** @type {string} */
  let tempDir;
  /** @type {string} */
  let githubOutputPath;
  /** @type {string} */
  let githubEnvPath;
  /** @type {Record<string, string|undefined>} */
  let originalEnv;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "log").mockImplementation(() => {});

    // Set up the mock to return valid IDs by default
    mockSendJobSetupSpan.mockResolvedValue({ traceId: VALID_TRACE_ID, spanId: VALID_SPAN_ID });
    sendOtlpModule.sendJobSetupSpan = mockSendJobSetupSpan;

    // Create temp files for GITHUB_OUTPUT and GITHUB_ENV
    tempDir = mkdtempSync(join(tmpdir(), "action-setup-otlp-test-"));
    githubOutputPath = join(tempDir, "github_output");
    githubEnvPath = join(tempDir, "github_env");
    writeFileSync(githubOutputPath, "");
    writeFileSync(githubEnvPath, "");

    originalEnv = {
      OTEL_EXPORTER_OTLP_ENDPOINT: process.env.OTEL_EXPORTER_OTLP_ENDPOINT,
      SETUP_START_MS: process.env.SETUP_START_MS,
      GITHUB_OUTPUT: process.env.GITHUB_OUTPUT,
      GITHUB_ENV: process.env.GITHUB_ENV,
      INPUT_TRACE_ID: process.env.INPUT_TRACE_ID,
      "INPUT_TRACE-ID": process.env["INPUT_TRACE-ID"],
      INPUT_JOB_NAME: process.env.INPUT_JOB_NAME,
      "INPUT_JOB-NAME": process.env["INPUT_JOB-NAME"],
    };

    delete process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
    delete process.env.SETUP_START_MS;
    delete process.env.INPUT_TRACE_ID;
    delete process.env["INPUT_TRACE-ID"];
    delete process.env.INPUT_JOB_NAME;
    delete process.env["INPUT_JOB-NAME"];
    process.env.GITHUB_OUTPUT = githubOutputPath;
    process.env.GITHUB_ENV = githubEnvPath;
  });

  afterEach(() => {
    vi.restoreAllMocks();
    sendOtlpModule.sendJobSetupSpan = originalSendJobSetupSpan;

    for (const [key, value] of Object.entries(originalEnv)) {
      if (value !== undefined) {
        process.env[key] = value;
      } else {
        delete process.env[key];
      }
    }

    // Clean up temp files
    if (existsSync(githubOutputPath)) unlinkSync(githubOutputPath);
    if (existsSync(githubEnvPath)) unlinkSync(githubEnvPath);
  });

  it("should export run as a function", () => {
    expect(typeof run).toBe("function");
  });

  describe("when OTEL_EXPORTER_OTLP_ENDPOINT is not set", () => {
    it("should log that the setup span is being skipped", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith("[otlp] OTEL_EXPORTER_OTLP_ENDPOINT not set, skipping setup span");
    });

    it("should still call sendJobSetupSpan", async () => {
      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledOnce();
    });

    it("should not log the 'setup span sent' message", async () => {
      await run();

      expect(console.log).not.toHaveBeenCalledWith(expect.stringContaining("setup span sent"));
    });
  });

  describe("when OTEL_EXPORTER_OTLP_ENDPOINT is set", () => {
    beforeEach(() => {
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318";
    });

    it("should call sendJobSetupSpan once", async () => {
      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledOnce();
    });

    it("should log the endpoint URL when sending", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith("[otlp] sending setup span to http://localhost:4318");
    });

    it("should log the setup span sent message with traceId and spanId", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith(`[otlp] setup span sent (traceId=${VALID_TRACE_ID}, spanId=${VALID_SPAN_ID})`);
    });

    it("should log the resolved trace-id", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith(`[otlp] resolved trace-id=${VALID_TRACE_ID}`);
    });
  });

  describe("INPUT_TRACE_ID handling", () => {
    it("should log that trace will be reused when INPUT_TRACE_ID is set", async () => {
      process.env.INPUT_TRACE_ID = VALID_TRACE_ID;

      await run();

      expect(console.log).toHaveBeenCalledWith(`[otlp] INPUT_TRACE_ID=${VALID_TRACE_ID} (will reuse activation trace)`);
    });

    it("should log that a new trace ID will be generated when INPUT_TRACE_ID is not set", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith("[otlp] INPUT_TRACE_ID not set, a new trace ID will be generated");
    });

    it("should pass INPUT_TRACE_ID to sendJobSetupSpan", async () => {
      process.env.INPUT_TRACE_ID = VALID_TRACE_ID;

      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledWith(expect.objectContaining({ traceId: VALID_TRACE_ID }));
    });

    it("should normalize INPUT_TRACE_ID to lowercase", async () => {
      process.env.INPUT_TRACE_ID = "ABCDEF1234567890ABCDEF1234567890";

      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledWith(expect.objectContaining({ traceId: "abcdef1234567890abcdef1234567890" }));
    });

    it("should pass traceId as undefined when INPUT_TRACE_ID is not set", async () => {
      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledWith(expect.objectContaining({ traceId: undefined }));
    });

    it("should accept INPUT_TRACE-ID (hyphen form)", async () => {
      process.env["INPUT_TRACE-ID"] = VALID_TRACE_ID;

      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledWith(expect.objectContaining({ traceId: VALID_TRACE_ID }));
    });
  });

  describe("INPUT_JOB_NAME handling", () => {
    it("should set INPUT_JOB_NAME env var when provided", async () => {
      process.env.INPUT_JOB_NAME = "agent";

      await run();

      expect(process.env.INPUT_JOB_NAME).toBe("agent");
    });
  });

  describe("SETUP_START_MS handling", () => {
    it("should pass startMs from SETUP_START_MS to sendJobSetupSpan", async () => {
      const startMs = Date.now() - 5000;
      process.env.SETUP_START_MS = String(startMs);

      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledWith(expect.objectContaining({ startMs }));
    });

    it("should default startMs to 0 when SETUP_START_MS is not set", async () => {
      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledWith(expect.objectContaining({ startMs: 0 }));
    });

    it("should default startMs to 0 when SETUP_START_MS is empty string", async () => {
      process.env.SETUP_START_MS = "";

      await run();

      expect(mockSendJobSetupSpan).toHaveBeenCalledWith(expect.objectContaining({ startMs: 0 }));
    });
  });

  describe("GITHUB_OUTPUT writing", () => {
    it("should write trace-id to GITHUB_OUTPUT when traceId is valid", async () => {
      await run();

      const contents = readFileSync(githubOutputPath, "utf8");
      expect(contents).toContain(`trace-id=${VALID_TRACE_ID}`);
    });

    it("should log that trace-id was written to GITHUB_OUTPUT", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith(`[otlp] trace-id=${VALID_TRACE_ID} written to GITHUB_OUTPUT`);
    });

    it("should not write to GITHUB_OUTPUT when traceId is invalid", async () => {
      mockSendJobSetupSpan.mockResolvedValue({ traceId: INVALID_TRACE_ID, spanId: INVALID_SPAN_ID });

      await run();

      const contents = readFileSync(githubOutputPath, "utf8");
      expect(contents).toBe("");
    });

    it("should not write to GITHUB_OUTPUT when GITHUB_OUTPUT is not set", async () => {
      delete process.env.GITHUB_OUTPUT;

      await run();

      // Should not throw, should just skip writing
      expect(mockSendJobSetupSpan).toHaveBeenCalledOnce();
    });
  });

  describe("GITHUB_ENV writing", () => {
    it("should write GITHUB_AW_OTEL_TRACE_ID to GITHUB_ENV when traceId is valid", async () => {
      await run();

      const contents = readFileSync(githubEnvPath, "utf8");
      expect(contents).toContain(`GITHUB_AW_OTEL_TRACE_ID=${VALID_TRACE_ID}`);
    });

    it("should write GITHUB_AW_OTEL_PARENT_SPAN_ID to GITHUB_ENV when spanId is valid", async () => {
      await run();

      const contents = readFileSync(githubEnvPath, "utf8");
      expect(contents).toContain(`GITHUB_AW_OTEL_PARENT_SPAN_ID=${VALID_SPAN_ID}`);
    });

    it("should always write GITHUB_AW_OTEL_JOB_START_MS to GITHUB_ENV", async () => {
      await run();

      const contents = readFileSync(githubEnvPath, "utf8");
      expect(contents).toMatch(/GITHUB_AW_OTEL_JOB_START_MS=\d+/);
    });

    it("should not write GITHUB_AW_OTEL_TRACE_ID when traceId is invalid", async () => {
      mockSendJobSetupSpan.mockResolvedValue({ traceId: INVALID_TRACE_ID, spanId: VALID_SPAN_ID });

      await run();

      const contents = readFileSync(githubEnvPath, "utf8");
      expect(contents).not.toContain("GITHUB_AW_OTEL_TRACE_ID=");
    });

    it("should not write GITHUB_AW_OTEL_PARENT_SPAN_ID when spanId is invalid", async () => {
      mockSendJobSetupSpan.mockResolvedValue({ traceId: VALID_TRACE_ID, spanId: INVALID_SPAN_ID });

      await run();

      const contents = readFileSync(githubEnvPath, "utf8");
      expect(contents).not.toContain("GITHUB_AW_OTEL_PARENT_SPAN_ID=");
    });

    it("should not write to GITHUB_ENV when GITHUB_ENV is not set", async () => {
      delete process.env.GITHUB_ENV;

      await run();

      // Should not throw, should just skip writing env vars
      expect(mockSendJobSetupSpan).toHaveBeenCalledOnce();
    });

    it("should log GITHUB_AW_OTEL_TRACE_ID written message", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith("[otlp] GITHUB_AW_OTEL_TRACE_ID written to GITHUB_ENV");
    });

    it("should log GITHUB_AW_OTEL_PARENT_SPAN_ID written message", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith("[otlp] GITHUB_AW_OTEL_PARENT_SPAN_ID written to GITHUB_ENV");
    });

    it("should log GITHUB_AW_OTEL_JOB_START_MS written message", async () => {
      await run();

      expect(console.log).toHaveBeenCalledWith("[otlp] GITHUB_AW_OTEL_JOB_START_MS written to GITHUB_ENV");
    });
  });

  describe("error handling", () => {
    it("should propagate errors from sendJobSetupSpan", async () => {
      mockSendJobSetupSpan.mockRejectedValueOnce(new Error("Network error"));

      await expect(run()).rejects.toThrow("Network error");
    });
  });
});
