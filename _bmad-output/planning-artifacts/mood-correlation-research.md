# Mood-Task Correlation Research

**Date:** 2026-03-02
**Story:** 15.2 — Mood-Task Correlation & Procrastination Research
**Purpose:** Evidence base for Epic 4's mood-aware adaptive door selection algorithm

---

## Executive Summary

This document synthesizes research on how mood states affect task performance, preference, and engagement. The findings inform ThreeDoors' adaptive algorithm (Epic 4, Stories 4.2–4.5), which uses mood data to select which tasks to present behind the three doors.

**Key finding:** Mood has well-documented, differential effects on task performance depending on task type. A two-dimensional mood model (valence × arousal) provides the minimum viable signal for meaningful task-mood matching.

---

## 1. Core Mood-Productivity Models

### 1.1 Yerkes-Dodson Law (1908)

Performance follows an inverted-U curve relative to arousal. The optimal arousal level depends on task complexity:

- **Simple/routine tasks:** optimal at higher arousal
- **Complex/cognitive tasks:** optimal at lower-to-moderate arousal

**Implication:** Stressed users (high arousal) should see simpler tasks. Bored users (low arousal) need moderate stimulation, not the most demanding options.

### 1.2 Broaden-and-Build Theory (Fredrickson, 1998, 2001)

Positive emotions widen attentional scope, promote exploration, and encourage creative, associative thinking. Negative emotions narrow the thought-action repertoire, focusing attention on specific threats.

**Implication:** Positive moods are the right time for creative, exploratory tasks. Negative moods favor focused, analytical work.

### 1.3 Affect Infusion Model (Forgas, 1995)

Mood influences cognition most strongly in open-ended processing situations. Key asymmetry:

- **Positive affect** → heuristic (fast, associative) processing
- **Negative affect** → substantive (slow, systematic, analytical) processing

**Implication:** Negative-mood users can be well-suited for detail-oriented tasks. Positive-mood users excel at association and pattern recognition but may rush judgments.

### 1.4 Feelings-as-Information Theory (Schwarz & Clore, 1983)

People use current affect as information when making judgments ("How do I feel about this?"). This operates even when mood is incidental (unrelated to the task). Users in negative moods perceive all tasks as more aversive than they actually are.

**Implication:** The system must account for mood-congruence bias. Task framing and presentation matter — especially in negative mood states.

### 1.5 Russell's Circumplex Model of Affect (1980)

All emotional states map onto two independent dimensions:

| Dimension | Poles |
|-----------|-------|
| **Valence** | Pleasure ↔ Displeasure |
| **Arousal** | High activation ↔ Low activation |

This produces four mood quadrants:

| Quadrant | Valence | Arousal | Examples |
|----------|---------|---------|----------|
| High positive activation | + | High | Excited, enthusiastic, alert |
| Low positive activation | + | Low | Content, calm, relaxed |
| High negative activation | − | High | Anxious, stressed, angry |
| Low negative activation | − | Low | Sad, fatigued, bored |

**Implication:** This is the most actionable mood capture framework for ThreeDoors. Capturing both valence and arousal provides richer matching signal than a single scale.

---

## 2. Mood-Task Type Matching Matrix

### Positive + High Arousal (Excited, Enthusiastic)

**Best for:** Creative/divergent thinking, brainstorming, strategic planning, collaborative tasks, exploration
**Avoid:** Detail checking, proofreading, final edits (risk of rushing)

*Evidence: Isen, Daubman & Nowicki, 1987; Fredrickson, 1998, 2001; Gasper & Clore, 2002*

### Positive + Low Arousal (Content, Calm)

**Best for:** Steady execution of known workflows, learning and knowledge consolidation, thoughtful writing
**Avoid:** High-urgency tasks

*Evidence: Fredrickson, 2001; Isen, 2001*

### Negative + High Arousal (Anxious, Stressed)

**Best for:** Low-demand completable tasks with clear endpoints, routine/administrative work, organizing
**Avoid:** Complex open-ended creative tasks, multi-step decisions, new learning

*Evidence: Yerkes & Dodson, 1908; Forgas, 1995; Amabile & Kramer, 2011*

Key insight: Amabile & Kramer's Progress Principle (2011, analysis of 12,000 diary entries from 238 workers) found that making progress on any meaningful work — even minor — was the strongest driver of positive inner work life. Offering completable tasks to stressed users generates mood repair via completion.

### Negative + Low Arousal (Sad, Fatigued)

**Best for:** Systematic/methodical analytical tasks, error-checking, proofreading, structured single-task work
**Avoid:** Creative tasks, socially demanding tasks, sustained-motivation tasks

*Evidence: Forgas, 1995; Gasper & Clore, 2002*

### Neutral

**Best for:** Moderate-complexity tasks, catch-up work, anything in the backlog
**Avoid:** Nothing strongly contraindicated — may be the best time to tackle avoided tasks (no mood bias)

---

## 3. Task Categorization Dimensions

For mood-based matching, tasks should be classified along these dimensions:

| Dimension | Scale | Rationale |
|-----------|-------|-----------|
| **Cognitive demand** | Low / Medium / High | Maps to Yerkes-Dodson arousal interaction |
| **Thinking mode** | Convergent / Divergent | Maps to positive/negative affect processing styles |
| **Energy requirement** | Low-energy compatible / Requires alertness | Filters by arousal level |
| **Social vs. Solo** | Social / Solo | Social tasks need positive affect |
| **Closure speed** | Quick win / Extended effort | Quick wins repair mood in stressed/fatigued states |
| **Stakes/pressure** | Low / High | High stakes + high stress = Yerkes-Dodson degradation |

ThreeDoors' existing categorization (Story 4.1: type, effort, context) aligns well. The main addition needed is explicitly tagging cognitive demand and thinking mode.

---

## 4. Circadian Rhythm Effects

### General Patterns

- **Morning (9am–12pm):** Peak alertness, working memory, executive function. Best for complex analytical work.
- **Early afternoon (1pm–3pm):** Post-lunch dip. 7–40% performance decrements on attention tasks. Best for routine work.
- **Late afternoon (3pm–6pm):** Second wind — reaction time improves, procedural tasks peak.
- **Evening:** Mind-wandering increases.

### Chronotype Synchrony Effect

Cognitive performance is best when task timing aligns with individual chronotype peaks. Evening chronotypes show the same patterns, shifted by several hours (May et al., 2023, *Perspectives on Psychological Science*).

**Implication:** Time-of-day is a separate signal from mood, but the two interact — users report lower energy during circadian troughs. ThreeDoors could weight time-of-day as a prior modulating task recommendations.

---

## 5. Choice Architecture and Mood

### Three Options Is Well-Grounded

The Paradox of Choice (Schwartz, 2004) and Hick's Law both support limiting options. Users in negative moods or under stress are particularly vulnerable to option overload. ThreeDoors' three-door constraint is a protective feature.

### Mood Effects on Choice Behavior

- **Happy mood:** increases heuristic reliance, willingness to engage, risk tolerance
- **Sad mood:** promotes systematic evaluation but can trigger avoidance of all options
- **Anxious mood:** promotes risk aversion and preference for predictable outcomes

**Implication:** Under stress/anxiety, users are more likely to engage if at least one option is clearly achievable and low-risk. The algorithm should anchor the set with one "safe" task (Lerner et al., 2015, *Annual Review of Psychology*).

---

## 6. Mood Measurement Approaches

### Recommended: Two-Dimensional Capture

Capture both valence and arousal separately (Russell's model):

1. "How are you feeling?" (valence: 😫 😕 😐 🙂 😄)
2. "What's your energy level?" (arousal: 🔋→ low to high)

This maps user states to the four quadrants used in task matching. Research supports:

- **Single-item scales** have acceptable validity for within-person variation over time (which is what adaptive selection needs)
- **Emoji-based ratings** achieve comparable results to numerical Likert scales with lower friction and higher completion rates
- **EMA-style check-ins** (mood at session start) are the methodological gold standard for real-world mood measurement (Shiffman, Stone & Hufford, 2008)

### Existing ThreeDoors Implementation

ThreeDoors already captures mood via `mood_selector.go` with emoji-based selection. The current implementation maps to valence. Adding an energy/arousal dimension would unlock the full quadrant model.

---

## 7. Recommendations for Epic 4

### R1: Adopt Two-Dimensional Mood Model
Add arousal/energy capture alongside existing valence capture. This enables the full mood-task matching matrix without adding significant user friction.

### R2: Implement Mood-Task Matching Weights
Use the quadrant-based matching matrix (Section 2) as default weights in the adaptive algorithm (Story 4.3). Weights should be configurable.

### R3: Anchor with a Safe Option
When user mood is negative-high-arousal (stressed), ensure at least one of the three doors contains a quick-win, low-demand task.

### R4: Account for Mood-Congruence Bias
Users in negative moods perceive all tasks as more aversive. Consider positive framing of task descriptions when mood is low.

### R5: Use Time-of-Day as a Secondary Signal
Weight task cognitive demand by time of day, particularly during the post-lunch dip (1–3pm) when attentional performance drops.

### R6: Track Mood-Completion Correlations
Store mood at task selection alongside completion data to build a per-user model. The population-level defaults from this research should be refined with individual data over time.

---

## Bibliography

- Amabile, T. & Kramer, S. (2011). *The Progress Principle: Using Small Wins to Ignite Joy, Engagement, and Creativity at Work*. Harvard Business Review Press. [HBR article](https://hbr.org/2011/05/the-power-of-small-wins)
- Forgas, J.P. (1995). Mood and judgment: The affect infusion model (AIM). *Psychological Bulletin*, 117(1), 39–66. [PubMed](https://pubmed.ncbi.nlm.nih.gov/7870863/)
- Fredrickson, B.L. (2001). The role of positive emotions in positive psychology: The broaden-and-build theory of positive emotions. *American Psychologist*, 56(3), 218–226. [PMC](https://pmc.ncbi.nlm.nih.gov/articles/PMC1693418/)
- Gasper, K. & Clore, G.L. (2002). Attending to the big picture: Mood and global versus local processing of visual information. *Psychological Science*, 13(1), 34–40.
- Isen, A.M., Daubman, K.A. & Nowicki, G.P. (1987). Positive affect facilitates creative problem solving. *Journal of Personality and Social Psychology*, 52(6), 1122–1131. [PubMed](https://pubmed.ncbi.nlm.nih.gov/3598858/)
- Lerner, J.S., Li, Y., Valdesolo, P. & Kassam, K.S. (2015). Emotion and decision making. *Annual Review of Psychology*, 66, 799–823.
- May, C.P. et al. (2023). Chronotype and cognitive performance. *Perspectives on Psychological Science*.
- Russell, J.A. (1980). A circumplex model of affect. *Journal of Personality and Social Psychology*, 39(6), 1161–1178.
- Schwartz, B. (2004). *The Paradox of Choice: Why More Is Less*. Ecco/HarperCollins.
- Schwarz, N. & Clore, G.L. (1983). Mood, misattribution, and judgments of well-being: Informative and directive functions of affective states. *Journal of Personality and Social Psychology*, 45(3), 513–523.
- Shiffman, S., Stone, A.A. & Hufford, M.R. (2008). Ecological momentary assessment. *Annual Review of Clinical Psychology*, 4, 1–32.
- Watson, D., Clark, L.A. & Tellegen, A. (1988). Development and validation of brief measures of positive and negative affect: The PANAS scales. *Journal of Personality and Social Psychology*, 54(6), 1063–1070. [PubMed](https://pubmed.ncbi.nlm.nih.gov/3397865/)
- Yerkes, R.M. & Dodson, J.D. (1908). The relation of strength of stimulus to rapidity of habit-formation. *Journal of Comparative Neurology and Psychology*, 18(5), 459–482.
