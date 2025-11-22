<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Before & After: Conversation Flow Improvements

This document shows how the skill's conversation flow has improved to be more natural and less overwhelming.

---

## Creating a Prompt

### ❌ Before: Question Overload

```
You: "Help me create a prompt for data analysis"

Claude: "I'd be happy to help! To create an effective prompt, I need to understand:

1. What task should the prompt accomplish?
2. Who is the intended user/audience?
3. What does good output look like?
4. Are there constraints (length, format, tone)?
5. Do you have examples of desired input/output?
6. What type of data will be analyzed?
7. What insights are you looking for?
8. What should the output format be?

Please provide as much detail as possible."

You: 😰 [overwhelmed, provides partial answers]

Claude: "Thanks! Just a few more questions..."
[More questions...]
```

**Problems:**
- 8+ questions at once
- Intimidating wall of text
- User can't remember all questions
- Kills conversational momentum

---

### ✅ After: Conversational Flow

**Quick Mode (Default):**
```
You: "Help me create a prompt for data analysis"

Claude: "I'll create a data analysis prompt. What type of data - sales, user behavior, or financial?"

You: "Sales data"

Claude: [creates prompt immediately]

"I've assumed quarterly sales analysis with trend identification. Let me know if you need different focus!"

You: "Add competitor comparison"

Claude: [updates prompt]

Done!
```

**Guided Mode:**
```
You: "Help me think through a prompt for data analysis"

Claude: "Let's explore this together. First, what type of data are you working with?"

You: "Sales data"

Claude: "Got it. What insights matter most - trends over time, product performance, or regional differences?"

You: "Trends and product performance"

Claude: [creates targeted prompt]

"Here's your prompt focused on temporal trends and product analysis."
```

**Expert Mode:**
```
You: "Quick prompt for sales data analysis"

Claude: [creates complete prompt immediately]

"Focused on trends and top performers. Ready to iterate!"
```

**Benefits:**
- 0-1 questions per turn
- Natural conversation flow
- User stays engaged
- Fast time-to-value

---

## Auditing a Prompt

### ❌ Before: Over-Detailed Analysis

```
You: "Audit this prompt: [paste prompt]"

Claude: "Let me provide a comprehensive analysis...

## Prompt Analysis

**Prompt Type**: Analysis Prompt (from taxonomy reference)

**Automated Analysis Results**:
============================================================
PROMPT ANALYSIS RESULTS
============================================================

Basic Metrics:
  - Length: 347 characters
  - Word count: 58 words
  - Uses XML tags: No
  - Has examples: No
  - Has role definition: Yes
  - Specifies output format: No
  - Uses thinking tags: No

⚠️  Specificity Issues (3):
  - \"concise/brief\" - Specify exact length (e.g., \"2-3 sentences\")
  - \"detailed\" - Specify what aspects to detail
  - \"good/better/best\" - Define quality criteria

⚠️  Best Practice Gaps (2):
  - No examples provided - Examples significantly improve output quality
  - No output format specified - Specify desired structure/format
============================================================

**Best Practice Assessment**:

✅ Strengths:
- Clear role definition establishes domain expertise
- Task is well-defined and specific
- Uses affirmative instructions

⚠️ Issues Found:
- Missing XML structure for organization (Claude was trained with XML)
- No examples provided to demonstrate expected quality
- Output format not specified, may lead to inconsistent results
- Uses vague terms like \"concise\" and \"detailed\" without specification
- Could benefit from thinking tags for transparent reasoning
- No context about data sources or constraints

💡 Suggested Improvements:
1. Add XML tags to structure the prompt (e.g., <role>, <task>, <output_format>)
2. Replace \"concise\" with specific length (e.g., \"3-5 sentences, maximum 100 words\")
3. Add example input/output pairs showing desired analysis quality
4. Specify exact output structure (e.g., \"Provide analysis as: Executive Summary (2-3 sentences), Key Findings (3 bullet points), Recommendations (numbered list)\")
5. Include <thinking> tags instruction for complex reasoning
6. Add context section about data types and constraints

Would you like me to rewrite this prompt incorporating these improvements?"

You: 😵 [information overload, lost track of what matters]
```

**Problems:**
- Wall of text, hard to scan
- Full script output included (unnecessary detail)
- Multiple sections with repetitive information
- User can't quickly see what matters most
- Takes 30+ seconds to read and understand

---

### ✅ After: Scannable Feedback

```
You: "Audit this prompt: [paste prompt]"

Claude: "## Quick Assessment

This is a functional analysis prompt, but it could be more effective.

✅ **Working well**:
- Clear role and task definition
- Affirmative instructions

⚠️ **Could improve**:
- Missing XML structure → Add <role>, <task>, <output_format> tags
- Vague \"concise\" → Specify \"3-5 sentences, 100 words max\"
- No examples → Add 1-2 input/output pairs
- No output format → Define structure (summary, findings, recommendations)

Want me to rewrite it with these improvements?"

You: ✅ [Understands instantly, can respond]
```

**Benefits:**
- Scannable in 5 seconds
- Lead with the conclusion
- Specific, actionable fixes
- No unnecessary detail
- User can engage immediately

---

## Improving a Failed Prompt

### ❌ Before: Over-Analysis of Failure

```
You: "This prompt didn't work, fix it: [paste prompt]"

Claude: "I'll help you diagnose and fix this prompt. Let me gather some information:

1. What was the prompt trying to achieve?
2. What output did you get?
3. What output did you want?
4. Why did it fail? (wrong format, missed requirements, incorrect reasoning, etc.)
5. Were there any specific errors or unexpected behaviors?
6. Can you share an example of the problematic output?

Please provide as much detail as possible so I can identify the root cause."

You: [provides some context]

Claude: "Thank you. Let me analyze this...

## Failure Analysis

**What went wrong**:
[Long explanation of the root cause]

**Why it happened**:
[Technical explanation with references to prompt engineering theory]

**Automated Analysis**:
[Full script output pasted]

**Identified Issues**:
1. [Issue 1]: [Lengthy explanation of how it caused the failure, referencing best practices]
2. [Issue 2]: [Another lengthy explanation with examples]
3. [Issue 3]: [Yet another detailed explanation]

**Recommended Fixes**:
1. [Fix 1 with detailed rationale and before/after examples]
2. [Fix 2 with multiple paragraphs of explanation]
3. [Fix 3 with references to documentation]

**Example of improved version**:
[Shows a snippet, but not the complete fixed prompt]

Would you like me to rewrite the complete prompt with these fixes?"

You: [Still waiting to see the actual fixed prompt]
```

**Problems:**
- 6 questions upfront when user already said it failed
- Over-analysis of something that clearly didn't work
- User has to wait through multiple turns to get the fix
- Fix is buried under walls of explanation

---

### ✅ After: Quick Diagnosis and Fix

```
You: "This prompt didn't work, fix it: [paste prompt]"

Claude: "What output did you get vs. what you wanted?"

You: "It gave me bullet points but I needed a structured report"

Claude: "## What Went Wrong

Missing output format specification - Claude didn't know you wanted report structure.

**Key Issues**:
- No <output_format> section → Added structured template
- Vague \"summarize\" → Now specifies sections and length
- No examples → Added sample report

Here's the fixed version:

[Complete fixed prompt with proper structure]

**What I changed**:
- Added <output_format> with report template
- Specified \"3 sections: Summary (2-3 sentences), Analysis (3 paragraphs), Recommendations (5 bullets)\"
- Included example report showing expected format

Try this version!"

You: ✅ [Can test immediately]
```

**Benefits:**
- 1 question max (only if needed)
- Shows fix immediately
- Brief explanation of changes
- User can test right away
- Natural to iterate if still not perfect

---

## Iterative Refinement

### ❌ Before: Required Re-pasting

```
You: "Create a prompt for code review"

Claude: [creates prompt]

You: "Make it shorter"

Claude: "Could you please paste the prompt again so I can shorten it?"

You: [pastes entire prompt again]

Claude: [shows shortened version]

You: "Add examples"

Claude: "Could you paste the current version?"

You: 😤 [pastes again...]
```

**Problems:**
- Required re-pasting every iteration
- Broke conversation flow
- Tedious and frustrating
- Discouraged iteration

---

### ✅ After: Conversational Iteration

```
You: "Create a prompt for code review"

Claude: [creates prompt]

You: "Make it shorter"

Claude: [shows condensed version]

"Reduced to 200 words while keeping core review criteria."

You: "Add examples"

Claude: [adds examples to current version]

You: "More formal tone"

Claude: [adjusts tone]

You: "Perfect!"
```

**Benefits:**
- Zero re-pasting needed
- Natural conversation flow
- Encourages iteration
- Fast refinement

---

## Summary: Key Improvements

| Aspect | Before | After |
|--------|--------|-------|
| **Questions per turn** | 5-8+ questions | 0-2 questions max |
| **Create workflow** | Multiple question rounds | Immediate generation |
| **Audit feedback** | 500+ word analysis | 50-100 words, scannable |
| **Improve workflow** | Long diagnosis, then ask to fix | Show fix immediately |
| **Iteration** | Re-paste every time | Natural references |
| **Approval** | Required for everything | Smart defaults |
| **Time to value** | 3-5 turns minimum | Often 1 turn |
| **User experience** | Overwhelming | Conversational |

## Design Principles Applied

1. **Bias toward action**: Generate first, refine later
2. **Progressive disclosure**: Show what matters, hide details unless asked
3. **Smart defaults**: Make assumptions, let user correct
4. **Minimal friction**: Remove unnecessary approval gates
5. **Conversational flow**: Natural dialogue, not Q&A forms
6. **Respect user time**: Lead with conclusions, not process
7. **Support iteration**: Easy refinement without re-pasting

These principles transform the skill from a formal tool into a natural conversational partner.
