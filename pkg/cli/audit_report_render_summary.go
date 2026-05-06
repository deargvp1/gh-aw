package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/github/gh-aw/pkg/console"
)

// renderCompactRunSummary outputs a dense multi-line block at the top of the audit report,
// optimized for LLM tokenizer efficiency. It concentrates the most critical signals in a
// compact key=value encoded format, enabling quick orientation without scanning all sections.
//
// Line 1 – run identity:   run= status= dur= event= branch=
// Line 2 – token economics: et= in= out= cache= cost= $/turn turns= tok/turn tpm=
// Line 3 – tool activity:   tools=Ntypes/Mcalls  top=name:N,name:N,...
// Line 4 – signals:         net=R/ok/blk(%)  findings=crit:N/high:N/med:N/low:N/info:N  posture=  domain=
func renderCompactRunSummary(data AuditData) {
	ov := data.Overview
	m := data.Metrics

	// Line 1: identity
	statusStr := ov.Status
	if ov.Conclusion != "" && ov.Conclusion != ov.Status {
		statusStr = ov.Status + "/" + ov.Conclusion
	}
	dur := ov.Duration
	if dur == "" {
		dur = "-"
	}
	fmt.Fprintf(os.Stderr, "  run=%-16d  status=%-22s  dur=%-8s  event=%s/%s\n",
		ov.RunID, statusStr, dur, ov.Event, ov.Branch)

	// Line 2: token economics
	var tokParts []string
	if m.EffectiveTokens > 0 {
		tokParts = append(tokParts, fmt.Sprintf("et=%s", rawNumber(m.EffectiveTokens)))
	} else if m.TokenUsage > 0 {
		tokParts = append(tokParts, fmt.Sprintf("tokens=%s", rawNumber(m.TokenUsage)))
	}
	if data.FirewallTokenUsage != nil && data.FirewallTokenUsage.TotalRequests > 0 {
		tu := data.FirewallTokenUsage
		tokParts = append(tokParts, fmt.Sprintf("in=%s out=%s", rawNumber(tu.TotalInputTokens), rawNumber(tu.TotalOutputTokens)))
		if tu.CacheEfficiency > 0 {
			tokParts = append(tokParts, fmt.Sprintf("cache=%.0f%%", tu.CacheEfficiency*100))
		}
	}
	effectiveTokens := m.EffectiveTokens
	if effectiveTokens == 0 {
		effectiveTokens = m.TokenUsage
	}
	if m.EstimatedCost > 0 {
		tokParts = append(tokParts, fmt.Sprintf("cost=$%.4f", m.EstimatedCost))
		if m.Turns > 0 {
			tokParts = append(tokParts, fmt.Sprintf("$%.4f/turn", m.EstimatedCost/float64(m.Turns)))
		}
	}
	if m.Turns > 0 {
		tokParts = append(tokParts, fmt.Sprintf("turns=%d", m.Turns))
		if effectiveTokens > 0 {
			tokParts = append(tokParts, fmt.Sprintf("%s tok/turn", rawNumber(effectiveTokens/m.Turns)))
		}
	}
	tpm := 0.0
	if data.SessionAnalysis != nil && data.SessionAnalysis.TokensPerMinute > 0 {
		tpm = data.SessionAnalysis.TokensPerMinute
	} else if data.PerformanceMetrics != nil && data.PerformanceMetrics.TokensPerMinute > 0 {
		tpm = data.PerformanceMetrics.TokensPerMinute
	}
	if tpm > 0 {
		tokParts = append(tokParts, fmt.Sprintf("tpm=%.0f", tpm))
	}
	if len(tokParts) > 0 {
		fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(tokParts, "  "))
	}

	// Line 3: tool activity
	if len(data.ToolUsage) > 0 {
		totalCalls := 0
		for _, t := range data.ToolUsage {
			totalCalls += t.CallCount
		}
		topN := 3
		if len(data.ToolUsage) < topN {
			topN = len(data.ToolUsage)
		}
		topTools := make([]string, 0, topN)
		for i := 0; i < topN; i++ {
			t := data.ToolUsage[i]
			topTools = append(topTools, fmt.Sprintf("%s:%d", t.Name, t.CallCount))
		}
		fmt.Fprintf(os.Stderr, "  tools=%dtypes/%dcalls  top=%s\n",
			len(data.ToolUsage), totalCalls, strings.Join(topTools, ","))
	}

	// Line 4: network + findings + posture + domain
	var sigParts []string
	if data.FirewallAnalysis != nil && data.FirewallAnalysis.TotalRequests > 0 {
		fa := data.FirewallAnalysis
		blkPct := 0.0
		if fa.TotalRequests > 0 {
			blkPct = float64(fa.BlockedRequests) / float64(fa.TotalRequests) * 100
		}
		sigParts = append(sigParts, fmt.Sprintf("net=%dreq/%dok/%dblk(%.0f%%)",
			fa.TotalRequests, fa.AllowedRequests, fa.BlockedRequests, blkPct))
	}
	// Findings by severity
	counts := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 0}
	for _, f := range data.KeyFindings {
		if _, ok := counts[f.Severity]; ok {
			counts[f.Severity]++
		} else {
			counts["info"]++
		}
	}
	if len(data.KeyFindings) > 0 {
		sigParts = append(sigParts, fmt.Sprintf("findings=crit:%d/high:%d/med:%d/low:%d/info:%d",
			counts["critical"], counts["high"], counts["medium"], counts["low"], counts["info"]))
	}
	if data.BehaviorFingerprint != nil && data.BehaviorFingerprint.ActuationStyle != "" {
		sigParts = append(sigParts, "posture="+data.BehaviorFingerprint.ActuationStyle)
	}
	if data.TaskDomain != nil && data.TaskDomain.Name != "" {
		sigParts = append(sigParts, "domain="+data.TaskDomain.Name)
	}
	if len(sigParts) > 0 {
		fmt.Fprintf(os.Stderr, "  %s\n", sigParts[0])
		if len(sigParts) > 1 {
			fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(sigParts[1:], "  "))
		}
	}

	fmt.Fprintln(os.Stderr)
}

// rawNumber formats an integer without locale separators, which is friendlier for LLM tokenizers.
// Numbers at or above 1,000,000 fall through to console.FormatNumber for readability.
func rawNumber(n int) string {
	if n >= 1_000_000 {
		return console.FormatNumber(n)
	}
	return fmt.Sprintf("%d", n)
}
