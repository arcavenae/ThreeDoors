# ThreeDoors Epic 1 Validation Decision Rubric

**Purpose:** Structured framework for evaluating Epic 1 success and deciding whether to proceed to Epic 2 (Apple Notes integration) or pivot/abandon.

**Decision Maker:** Product Owner (with developer input)

**Timeline:** Complete after 1 week of daily ThreeDoors usage

---

## Validation Period Summary

**Start Date:** _________________
**End Date:** _________________
**Total Days:** ______
**Total Sessions:** ______
**Unique Days Used:** ______

---

## Section 1: Quantitative Metrics (50% of Decision Weight)

### Metric 1: Time to First Door (Friction Metric)

**Definition:** Average seconds from app launch to selecting first door (pressing 1/2/3)

**Target:** < 30 seconds
**Measured:** ______ seconds

**Scoring:**
- ✅ **PASS (5 points):** < 30 seconds → Faster than scrolling list, low friction
- ⚠️ **MARGINAL (3 points):** 30-60 seconds → Comparable to list, unclear advantage
- ❌ **FAIL (0 points):** > 60 seconds → Slower than list, adds friction

**Score:** ______ / 5

**Notes:**
```
[Developer observations about decision-making speed, hesitation patterns, etc.]
```

---

### Metric 2: Refresh Usage Rate

**Definition:** Percentage of sessions where refresh (R) was used at least once

**Target:** ≥ 30%
**Measured:** ______%

**Scoring:**
- ✅ **PASS (5 points):** ≥ 30% → Refresh option is valuable, demonstrates choice flexibility
- ⚠️ **MARGINAL (3 points):** 20-29% → Some value, but underutilized
- ❌ **FAIL (0 points):** < 20% → Refresh not valuable, just random selection

**Score:** ______ / 5

**Notes:**
```
[When was refresh used? What prompted it? Was it helpful?]
```

---

### Metric 3: Task Completion Rate

**Definition:** Average tasks completed per session

**Target:** ≥ 1.0 tasks/session (flexible based on task sizing)
**Measured:** ______ tasks/session

**Scoring:**
- ✅ **PASS (5 points):** ≥ 1.0 → Demonstrates productivity
- ⚠️ **MARGINAL (3 points):** 0.5-0.99 → Some progress, check task sizing
- ❌ **FAIL (0 points):** < 0.5 → Tool not enabling progress

**Score:** ______ / 5

**Notes:**
```
[Task sizing observations, completion barriers, etc.]
```

---

### Metric 4: Usage Consistency

**Definition:** Number of unique days the app was used during validation period

**Target:** ≥ 5 days (out of 7)
**Measured:** ______ days

**Scoring:**
- ✅ **PASS (5 points):** ≥ 5 days → Habit formation, daily utility demonstrated
- ⚠️ **MARGINAL (3 points):** 3-4 days → Some adoption, extend validation period
- ❌ **FAIL (0 points):** < 3 days → Not sticky, validation insufficient

**Score:** ______ / 5

**Notes:**
```
[Why were some days skipped? What prevented usage?]
```

---

### Metric 5: Detail View Engagement

**Definition:** Percentage of door selections that led to detail view entry

**Target:** ≥ 50% (demonstrates task detail workflow value)
**Measured:** ______%

**Scoring:**
- ✅ **PASS (3 points):** ≥ 50% → Detail view adds value
- ⚠️ **MARGINAL (2 points):** 30-49% → Mixed utility
- ❌ **FAIL (0 points):** < 30% → Detail view not useful

**Score:** ______ / 3

**Notes:**
```
[What drove detail view usage vs. quick completion?]
```

---

**Section 1 Total:** ______ / 23 points

**Section 1 Grade:**
- 20-23 points: EXCELLENT ✅
- 15-19 points: GOOD ⚠️
- 10-14 points: MARGINAL ⚠️
- < 10 points: POOR ❌

---

## Section 2: Qualitative Assessment (30% of Decision Weight)

### Question 1: Cognitive Load & Decision-Making

**Prompt:** Compared to your previous task management approach, does Three Doors reduce cognitive load when deciding what to work on?

**Response:**
```
[5-10 sentences describing decision-making experience]
```

**Rating:**
- ✅ **5 points:** Significantly reduces cognitive load, feels effortless
- ⚠️ **3 points:** Somewhat reduces load, but still requires thinking
- ❌ **0 points:** Doesn't reduce load or adds complexity

**Score:** ______ / 5

---

### Question 2: User Experience & Flow

**Prompt:** Does the Three Doors interface help you get into "flow state" faster, or does it interrupt your workflow?

**Response:**
```
[Describe flow experience, interruptions, smoothness]
```

**Rating:**
- ✅ **5 points:** Enables faster flow, minimal interruption
- ⚠️ **3 points:** Neutral, neither helps nor hinders
- ❌ **0 points:** Interrupts flow, adds friction

**Score:** ______ / 5

---

### Question 3: Emotional Response

**Prompt:** How does the app make you feel? Motivated? Overwhelmed? Neutral? Does the "progress over perfection" philosophy come through?

**Response:**
```
[Emotional reaction, motivation levels, messaging impact]
```

**Rating:**
- ✅ **5 points:** Positive emotions, motivating, philosophy evident
- ⚠️ **3 points:** Neutral, functional but not inspiring
- ❌ **0 points:** Negative emotions, demotivating, stressful

**Score:** ______ / 5

---

### Question 4: Comparison to Alternatives

**Prompt:** Would you choose to use ThreeDoors over your previous task management approach if both were available?

**Response:**
```
[Direct comparison, trade-offs, preference reasoning]
```

**Rating:**
- ✅ **5 points:** Strong preference for ThreeDoors
- ⚠️ **3 points:** Slight preference or context-dependent
- ❌ **0 points:** Prefer previous approach

**Score:** ______ / 5

---

### Question 5: Pain Points & Frustrations

**Prompt:** What are the biggest pain points or frustrations you experienced? Are they fundamental to the Three Doors concept or implementation issues?

**Response:**
```
[Pain points, categorize as concept vs. implementation issues]
```

**Impact Assessment:**
- ✅ **5 points:** Only minor implementation issues, concept sound
- ⚠️ **3 points:** Some concept concerns but addressable
- ❌ **0 points:** Fundamental concept flaws

**Score:** ______ / 5

---

### Question 6: Feature Completeness

**Prompt:** Did the Tech Demo have enough functionality to validate the concept, or were critical features missing?

**Response:**
```
[Missing features that prevented proper validation]
```

**Rating:**
- ✅ **3 points:** Sufficient for validation
- ⚠️ **2 points:** Mostly sufficient, minor gaps
- ❌ **0 points:** Critical gaps prevent validation

**Score:** ______ / 3

---

**Section 2 Total:** ______ / 28 points

**Section 2 Grade:**
- 24-28 points: EXCELLENT ✅
- 18-23 points: GOOD ⚠️
- 12-17 points: MARGINAL ⚠️
- < 12 points: POOR ❌

---

## Section 3: Technical Assessment (20% of Decision Weight)

### Criteria 1: Architecture Readiness for Epic 2

**Question:** Is the current architecture (monolithic, direct dependencies) suitable for refactoring to adapter pattern in Epic 2?

**Assessment:**
```
[Code quality, refactoring ease, technical debt]
```

**Rating:**
- ✅ **5 points:** Clean code, straightforward refactor path
- ⚠️ **3 points:** Some refactoring challenges but manageable
- ❌ **0 points:** Major refactor required, throwaway code

**Score:** ______ / 5

---

### Criteria 2: Performance & Stability

**Question:** Were there performance issues, crashes, or bugs that impacted validation?

**Assessment:**
```
[Bugs encountered, crashes, performance lag]
```

**Rating:**
- ✅ **5 points:** Stable, performant, no major issues
- ⚠️ **3 points:** Minor bugs, acceptable for Tech Demo
- ❌ **0 points:** Frequent crashes or unusable performance

**Score:** ______ / 5

---

### Criteria 3: Data Integrity

**Question:** Did tasks.yaml and completed.txt maintain data integrity? Any corruption or data loss?

**Assessment:**
```
[Data issues, file corruption, lost data]
```

**Rating:**
- ✅ **3 points:** Perfect data integrity
- ⚠️ **2 points:** Minor issues, no data loss
- ❌ **0 points:** Data corruption or loss occurred

**Score:** ______ / 3

---

**Section 3 Total:** ______ / 13 points

**Section 3 Grade:**
- 11-13 points: EXCELLENT ✅
- 8-10 points: GOOD ⚠️
- 5-7 points: MARGINAL ⚠️
- < 5 points: POOR ❌

---

## Overall Validation Score

| Section | Points Earned | Max Points | Weight | Weighted Score |
|---------|---------------|------------|--------|----------------|
| 1. Quantitative Metrics | ______ | 23 | 50% | ______ |
| 2. Qualitative Assessment | ______ | 28 | 30% | ______ |
| 3. Technical Assessment | ______ | 13 | 20% | ______ |
| **TOTAL** | | | **100%** | **______** |

**Calculation:**
- Section 1 Weighted = (Points / 23) × 50
- Section 2 Weighted = (Points / 28) × 30
- Section 3 Weighted = (Points / 13) × 20
- Total = Sum of weighted scores (out of 100)

---

## Decision Matrix

Based on overall weighted score:

### ✅ PROCEED TO EPIC 2 (Score: 75-100)

**Recommendation:** Three Doors hypothesis validated. Proceed to Epic 2 (Apple Notes integration).

**Next Steps:**
1. Conduct Apple Notes integration spike (evaluate 4 integration options)
2. Refactor to adapter pattern (TaskProvider interface)
3. Plan Epic 2 stories based on learnings
4. Consider addressing pain points from Section 2 during refactor

**Confidence Level:** HIGH - Data supports hypothesis

---

### ⚠️ CONDITIONAL PROCEED (Score: 60-74)

**Recommendation:** Mixed results. Consider one of the following:

**Option A: Proceed with Awareness**
- Epic 2 is feasible but address specific weaknesses
- Document concerns and improvement areas
- Set additional validation criteria for Epic 2

**Option B: Iterate on Epic 1**
- Extend validation period (2 more weeks)
- Address specific pain points from qualitative assessment
- Re-evaluate with fresh data

**Option C: Pivot UX**
- Three Doors concept has merit but needs adjustment
- Consider: Different door selection algorithm, different presentation, hybrid approach
- Prototype adjustments before Epic 2 investment

**Next Steps:**
1. Review which sections scored poorly
2. Determine if issues are fixable or fundamental
3. Make informed decision based on cost/benefit

**Confidence Level:** MODERATE - Some validation, some concerns

---

### ❌ PIVOT OR ABANDON (Score: < 60)

**Recommendation:** Three Doors hypothesis not validated. Consider alternatives.

**Option A: Pivot to Different UX**
- Three Doors might not be the right metaphor
- Consider: Smart single-task recommendation, traditional list with smart sorting, context-aware task surfacing
- Salvage learnings (architecture, task model) but redesign UX

**Option B: Abandon Project**
- Concept doesn't reduce friction for this user
- Traditional tools may be sufficient
- Stop further investment

**Next Steps:**
1. Analyze what didn't work (root cause)
2. Determine if salvageable with major changes
3. Consider sunk cost vs. future investment

**Confidence Level:** LOW - Hypothesis not supported

---

## Supporting Evidence Summary

### Key Wins (What Worked Well)
```
1.
2.
3.
```

### Key Failures (What Didn't Work)
```
1.
2.
3.
```

### Surprising Insights (Unexpected Learnings)
```
1.
2.
3.
```

---

## Final Decision

**Date:** _________________

**Decision:**
- [ ] Proceed to Epic 2
- [ ] Conditional Proceed (specify option: ________)
- [ ] Pivot to Different UX
- [ ] Abandon Project

**Rationale (2-3 paragraphs):**
```








```

**Commitments (if proceeding):**
```
1.
2.
3.
```

**Sign-off:**

Product Owner: _________________ Date: _________

Developer: _________________ Date: _________

---

## Appendix: Raw Metrics Data

**Paste output from validation scripts:**

### analyze_sessions.sh output:
```
[Paste here]
```

### daily_completions.sh output:
```
[Paste here]
```

### validation_decision.sh output:
```
[Paste here]
```

---

*This rubric provides a structured, objective framework for making the Epic 1 → Epic 2 decision while balancing quantitative metrics with qualitative human experience.*
