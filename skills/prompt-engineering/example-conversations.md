<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Example Conversations: Three Modes in Action

This document shows realistic conversations using the prompt engineering skill in all three modes.

---

## Example 1: Quick Mode (Default)

**Scenario**: Need a code review prompt for a Python project

```
You: "Create a prompt for Python code review"

Claude: "I'll create a Python code review prompt. Should this focus on general best practices or specific concerns like security, performance, or style?"

You: "Security and best practices"

Claude: 
<role>You are an expert Python security reviewer and code quality specialist.</role>

<context>
You're reviewing Python code for security vulnerabilities and adherence to best practices.
Focus on identifying issues that could lead to security problems or maintainability concerns.
</context>

<task>
Review the provided Python code and identify:
1. Security vulnerabilities (SQL injection, XSS, authentication issues, etc.)
2. Best practice violations (PEP 8, error handling, code organization)
3. Potential bugs or edge cases
</task>

<instructions>
Use <thinking> tags to show your review process:
1. Scan for security patterns (user input handling, authentication, data validation)
2. Check code structure and error handling
3. Identify edge cases and potential failures
4. Prioritize findings by severity
</instructions>

<code>
[Paste Python code here]
</code>

<output_format>
## Security Issues
- **Critical**: [Issue with specific line numbers and fix]
- **Medium**: [Issue with specific line numbers and fix]
- **Low**: [Issue with specific line numbers and fix]

## Best Practice Violations
- [Issue with recommendation]
- [Issue with recommendation]

## Recommendations
1. [Specific, actionable fix with code example]
2. [Specific, actionable fix with code example]
</output_format>

---

**Key points**: Security-focused with severity levels, includes thinking tags for thorough analysis.

You: "Add example vulnerabilities"

Claude: [Updates prompt after <instructions> section]

<examples>
<example>
<vulnerable_code>
user_input = request.args.get('id')
query = f"SELECT * FROM users WHERE id = {user_input}"
cursor.execute(query)
</vulnerable_code>

<good_review>
**Critical**: SQL Injection vulnerability (line 2)
- User input directly interpolated into SQL query
- **Fix**: Use parameterized query: `cursor.execute("SELECT * FROM users WHERE id = ?", (user_input,))`
</good_review>
</example>
</examples>

Updated! Now includes examples of what good security reviews look like.

You: "Perfect, thanks!"
```

**Time**: 2 turns, ~30 seconds
**Questions**: 1 clarifying question
**Iterations**: 1 refinement

---

## Example 2: Guided Mode

**Scenario**: Not sure what's needed for customer support bot

```
You: "Help me think through a prompt for a customer support chatbot"

Claude: "Let's build this together. First, what products or services will this bot support?"

You: "SaaS product - project management software"

Claude: "Got it. Should the bot handle technical troubleshooting, billing questions, or general guidance?"

You: "All three, but billing should redirect to humans"

Claude: "Makes sense. What tone - professional and formal, or friendly and conversational?"

You: "Friendly but professional"

Claude: "Perfect. Last question - any specific constraints on response length or must-know company policies?"

You: "Keep responses short, 2-3 sentences usually. We have a 30-day refund policy that should be mentioned for billing."

Claude: 
<role>
You are a friendly customer support agent for TaskFlow, a project management SaaS platform.

Personality traits:
- Helpful and patient
- Knowledgeable about the product
- Professional but warm
- Solution-oriented
</role>

<context>
TaskFlow offers project management features including task tracking, team collaboration, and reporting.
Company policies:
- 30-day money-back guarantee on all plans
- 24/7 technical support available
- Billing issues should be escalated to human support team
</context>

<guidelines>
Response style:
- Keep responses 2-3 sentences, conversational tone
- Acknowledge customer's concern first
- Offer specific solutions, not vague promises
- If you can't help, explain why and provide alternative (e.g., billing issues → human support)

Boundaries:
- Don't make promises outside stated policies
- Don't process refunds or billing changes (direct to support team)
- If technical issue is complex, offer to escalate
- For feature requests, acknowledge and explain feedback process
</guidelines>

<examples>
<example>
Customer: "I can't figure out how to assign tasks to team members"
You: "I'd be happy to help with task assignment! In your project view, click the task you want to assign, then click 'Assign to' in the right panel and select your team member. If you don't see team members listed, make sure they've accepted their invitations in Settings → Team."
</example>

<example>
Customer: "Why was I charged $99 instead of $49?"
You: "I understand billing concerns can be stressful - I want to make sure this gets resolved quickly. I'll need to connect you with our billing team who can review your account and charges. Would you like me to create a priority ticket, or would you prefer to reach them directly at billing@taskflow.com?"
</example>

<example>
Customer: "This app is too slow!"
You: "I'm sorry you're experiencing slow performance - that's definitely frustrating when you're trying to work. Let me check: are you seeing slowness throughout the app, or mainly when loading large projects? Also, this might help immediately: try clearing your browser cache (Settings → Privacy → Clear Data) as this resolves speed issues for many users."
</example>
</examples>

<edge_cases>
- Angry customers: Acknowledge frustration, focus on solutions, don't take it personally
- Off-topic questions: Politely redirect to TaskFlow-related topics
- Competitor comparisons: Focus on TaskFlow's strengths without disparaging others
- Feature requests: "That's great feedback! I've noted your request for [feature]. Our product team reviews all suggestions."
</edge_cases>

---

**Key features**: Detailed persona, clear boundaries, multiple examples showing different scenarios, explicit escalation paths.

You: "This looks great! Can you make the billing redirect more prominent?"

Claude: [Updates guidelines section]

<guidelines>
Response style:
- Keep responses 2-3 sentences, conversational tone
- Acknowledge customer's concern first
- Offer specific solutions, not vague promises
- **For ANY billing questions**: Immediately redirect to human support with empathy

Boundaries:
- **BILLING**: Never attempt to solve billing issues - always redirect to support@taskflow.com or live chat
- Don't make promises outside stated policies
- If technical issue is complex, offer to escalate
- For feature requests, acknowledge and explain feedback process
</guidelines>

Made billing redirection more explicit and moved to top priority. Now it's clear this should be the immediate response for billing questions.

You: "Perfect, thank you!"
```

**Time**: 5 turns, ~2 minutes
**Questions**: 4 questions (1-2 per turn)
**Iterations**: 1 refinement
**Result**: More comprehensive prompt because we explored the requirements

---

## Example 3: Expert Mode

**Scenario**: You know exactly what you need

```
You: "Quick prompt for summarizing meeting notes - extract action items, decisions, and next steps"

Claude:
<role>You are an executive assistant specializing in meeting documentation.</role>

<task>
Summarize this meeting transcript and extract actionable information.
</task>

<meeting_transcript>
[Paste transcript here]
</meeting_transcript>

<output_format>
## Meeting Summary
[2-3 sentences capturing main outcomes]

## Key Decisions
- [Decision 1] - Decided by: [Name]
- [Decision 2] - Decided by: [Name]

## Action Items
1. [Task] - Owner: [Name] - Deadline: [Date]
2. [Task] - Owner: [Name] - Deadline: [Date]

## Open Questions
- [Question requiring follow-up]

## Next Meeting
[Date and time if mentioned, or "TBD"]
</output_format>

<instructions>
- Focus on decisions and actions, not discussion process
- Extract owner names and deadlines from transcript
- Omit small talk and tangential discussions
- If conflicting info discussed, note the final decision
- Keep summary under 300 words
</instructions>

---

Ready to use! Focused on actionable extraction with clear structure.

You: "Add parking lot items section"

Claude: [Updates output_format]

<output_format>
## Meeting Summary
[2-3 sentences capturing main outcomes]

## Key Decisions
- [Decision 1] - Decided by: [Name]
- [Decision 2] - Decided by: [Name]

## Action Items
1. [Task] - Owner: [Name] - Deadline: [Date]
2. [Task] - Owner: [Name] - Deadline: [Date]

## Open Questions
- [Question requiring follow-up]

## Parking Lot
- [Topic tabled for future discussion]
- [Idea mentioned but not acted upon]

## Next Meeting
[Date and time if mentioned, or "TBD"]
</output_format>

Added! Now captures ideas that were mentioned but deferred.

You: "Great!"
```

**Time**: 2 turns, ~20 seconds
**Questions**: 0
**Iterations**: 1 refinement

---

## Example 4: Audit Workflow

**Scenario**: Review before using

```
You: "Audit this prompt before I use it: Write a detailed analysis of the data trends."

Claude: "## Quick Assessment

This prompt is too vague to produce consistent, useful results.

✅ **Working well**:
- Clear task (analysis)

⚠️ **Critical issues**:
- No role or context
- "Detailed" is vague → Specify: "3 paragraphs, 300-400 words" or "5 key findings with supporting data"
- No output structure → Add format expectations
- Missing data reference → Where should Claude look for the data?
- No analysis methodology specified

Want me to rewrite this as a proper analysis prompt?"

You: "Yes, make it for sales data analysis"

Claude:
<role>You are a business intelligence analyst specializing in sales data.</role>

<context>
Analyzing sales performance data to identify trends and actionable insights.
</context>

<task>
Analyze the provided sales data and identify key trends, patterns, and opportunities.
</task>

<data>
[Insert your sales data here - CSV, JSON, or table format]
</data>

<methodology>
Use <thinking> tags to show your analysis:
1. Calculate key metrics (growth rates, market share, trends)
2. Identify patterns across time periods and categories
3. Flag anomalies or unexpected changes
4. Draw actionable conclusions
</methodology>

<output_format>
## Executive Summary
[2-3 sentences with key takeaway]

## Key Trends
1. [Trend with supporting data and percentage changes]
2. [Trend with supporting data and percentage changes]
3. [Trend with supporting data and percentage changes]

## Opportunities
- [Actionable opportunity based on data]
- [Actionable opportunity based on data]

## Concerns
- [Risk or declining metric to watch]
- [Risk or declining metric to watch]

## Recommendations
1. [Specific action with rationale]
2. [Specific action with rationale]
</output_format>

<constraints>
- Use specific numbers and percentages
- Compare to previous periods where possible
- Focus on actionable insights, not just observations
- Limit analysis to 400-500 words
</constraints>

---

**What I changed**:
- Added role, context, and clear data section
- Replaced "detailed" with specific 400-500 word limit
- Added structured output format with required sections
- Included methodology with thinking tags for transparency
- Specified constraints for consistency

You: "Much better, thanks!"
```

**Time**: 2 turns
**Questions**: 0 (had enough context from the request)

---

## Example 5: Improve Workflow

**Scenario**: Prompt isn't working

```
You: "This prompt isn't working - it keeps giving me bullet points when I need prose: Explain the water cycle."

Claude: "## What Went Wrong

Missing output format specification - Claude defaulted to bullets when you wanted paragraphs.

**Key Issues**:
- No format specified → Added prose requirement
- No depth guidance → Added 3-paragraph structure
- Too simple → Added role and context for better quality

Here's the fixed version:

<role>You are a science educator writing for middle school students.</role>

<task>
Explain the water cycle in clear, engaging prose.
</task>

<output_format>
Provide your explanation as 3 paragraphs:
1. Overview of what the water cycle is and why it matters (3-4 sentences)
2. Description of the main stages: evaporation, condensation, precipitation (5-6 sentences with examples)
3. The cycle's importance to life on Earth (3-4 sentences)

Write in complete paragraphs, not bullet points or lists.
</output_format>

<tone>
Clear and engaging, using analogies that middle school students can relate to.
</tone>

<instructions>
- Use everyday examples (puddles drying, clouds forming, rain falling)
- Avoid jargon; if technical terms are needed, define them simply
- Include at least one analogy to make it memorable
</instructions>

---

**What I changed**:
- Added explicit "3 paragraphs" requirement
- Specified "not bullet points or lists"
- Added role for appropriate reading level
- Included tone guidance for engaging writing

Try this version!

You: "Perfect, that works!"
```

**Time**: 1 turn
**Question**: 0 (user explained the failure)

---

## Key Takeaways

### Quick Mode
- **Best for**: 80% of use cases
- **Questions**: 0-1
- **Speed**: Fastest
- **Iterations**: Often needed to refine

### Guided Mode
- **Best for**: Exploration, uncertain requirements
- **Questions**: 1-2 per turn, 3-5 turns total
- **Speed**: Slower but more deliberate
- **Iterations**: Fewer needed

### Expert Mode
- **Best for**: Clear vision, experienced users
- **Questions**: 0
- **Speed**: Instant
- **Iterations**: Minimal refinement

### Universal Benefits
- No re-pasting ever
- Natural conversation flow
- Smart defaults with easy correction
- Immediate generation in most cases
- Easy iteration and refinement

All modes support the same quality of prompts - they just differ in how you get there!
