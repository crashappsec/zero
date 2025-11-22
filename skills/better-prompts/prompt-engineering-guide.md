<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Prompt Engineering Skill - User Guide

## Conversation Modes

The skill adapts to your working style with three modes:

### 🚀 Quick Mode (Default)
**Best for**: Getting started fast and iterating as needed

- Asks only 1 essential question (or none if context is clear)
- Makes reasonable assumptions and states them
- Generates prompts immediately
- Refines through iteration

**Example:**
```
You: "Create a prompt for code review"
Claude: "I'll create a code review prompt. Python-focused or language-agnostic?"
You: "Python"
Claude: [creates complete prompt immediately]
```

### 🤔 Guided Mode
**Best for**: Exploring requirements thoughtfully

- Asks 1-2 questions per turn (never overwhelming lists)
- Builds understanding progressively
- More deliberate, less iteration needed

**Triggers**: "help me think through", "walk me through", "let's explore", "I'm not sure"

**Example:**
```
You: "Help me think through a customer support prompt"
Claude: "Let's build this together. What tone - friendly or professional?"
You: "Friendly"
Claude: "Got it. Should it handle complaints or focus on general questions?"
[continues thoughtfully]
```

### ⚡ Expert Mode
**Best for**: When you know exactly what you want

- Zero questions unless critically ambiguous
- Generates immediately from your description
- Fast, iterative refinement

**Triggers**: "quick prompt for", "just create", "make me a prompt"

**Example:**
```
You: "Quick prompt for meeting summaries"
Claude: [creates prompt immediately with smart defaults]
```

The mode is automatically detected from your language style - quick/direct → Expert, detailed → Quick, uncertain → Guided.

## How It Works

### No Question Overload

Gone are the days of:
```
❌ OLD WAY:
Claude: "To create this prompt, I need to know:
• What task should it accomplish?
• Who is the audience?
• What output format?
• What tone should it have?
• Are there constraints?
• Do you have examples?"
```

Now it's:
```
✅ NEW WAY:
You: "Create a prompt for data analysis"
Claude: "I'll create a data analysis prompt. What type of data - sales, user behavior, or financial?"
You: "Sales"
Claude: [creates prompt with smart assumptions]
```

### Smart Assumptions

Claude makes reasonable assumptions and lets you correct them:
```
You: "Create a prompt for API documentation"
Claude: [creates prompt]
"I've assumed REST APIs with JSON responses. Let me know if that needs adjusting!"
```

## What You've Got

Your new **prompt-engineering** skill is an expert assistant that helps you create, audit, and improve prompts optimized for Claude. It includes:

### Core Components

1. **SKILL.md** - Main workflow that guides Claude through three modes:
   - **Audit Mode**: Review existing prompts and suggest improvements
   - **Create Mode**: Help you write new prompts from scratch
   - **Improve Mode**: Fix prompts that didn't work as expected

2. **Automated Analysis Script** (`analyze_prompt.py`)
   - Scans prompts for vague language and missing elements
   - Identifies best practice gaps automatically
   - Provides structured feedback

3. **Comprehensive References**:
   - **taxonomy.md**: 8 prompt categories (Technical, Analysis, Creative, etc.) with characteristics and patterns
   - **best-practices.md**: Claude-specific techniques, common pitfalls, optimization strategies
   - **examples.md**: Ready-to-use example prompts for each category

## How to Use It

### Installation

1. Go to claude.ai and navigate to your project
2. Click the tools icon and select "Add Skill"
3. Upload the `prompt-engineering.skill` file
4. The skill is now available for use

### Usage Patterns

**To audit a prompt:**
```
"Review this prompt: [paste your prompt]"
"Is this prompt good? [paste prompt]"
```

**Or work conversationally:**
```
[You create a prompt in chat]
"Audit that prompt"
"Review the prompt above"
"Check the last prompt I wrote"
```

**To create a new prompt:**
```
"Help me write a prompt to [describe your goal]"
"I need a prompt that can [describe what you want]"
"Create a prompt for [task description]"
```

**To improve and iterate:**
```
"This prompt didn't work: [paste prompt]"
"Improve the last prompt"
"Make it more specific"
"Add examples to that prompt"
"Change the tone to be more formal"
"Shorten it"
```

### Conversational Flow (No Re-pasting!)

The skill works naturally with conversation context:

```
You: "Create a prompt for code review"
Claude: [generates prompt]

You: "Add examples"
Claude: [adds examples to that prompt]

You: "Make it focus on security"
Claude: [refines to emphasize security]

You: "Now audit it"
Claude: [analyzes the current version]

You: "Perfect, make it shorter"
Claude: [condenses while preserving key elements]
```

**No need to re-paste the prompt** - just refer to it naturally ("that prompt", "the last one", "it")

### What to Expect

The flow depends on what you're doing:

**Creating a prompt:**
1. **Quick/Expert Mode**: Gets prompt immediately (or after 1 question)
2. **Guided Mode**: 2-4 turns with 1-2 questions each
3. Ready to iterate naturally

**Auditing a prompt:**
1. Quick assessment: "Solid but could be more specific"
2. Scannable feedback (3-5 bullets)
3. Asks if you want it rewritten
4. Rewrites after your approval

**Improving a failed prompt:**
1. Asks 1 question if needed ("what went wrong?")
2. Shows fixed version immediately
3. Brief note on what changed (2-3 bullets)
4. Ready to iterate

**Iterating:**
- "Make it shorter" → Updated version + one-line note
- "Add examples" → Updated with examples
- "More formal" → Adjusted tone
- No re-approval needed for refinements

## Key Features

### Extensible Taxonomy

The skill comes with 8 pre-defined prompt categories:
- Technical (code, debugging, architecture)
- Analysis (data analysis, research, critical thinking)
- Creative (writing, content creation, brainstorming)
- Instructional (tutorials, explanations, documentation)
- Conversation/Dialogue (chatbots, customer service)
- Summarization (condensing content, key points)
- Classification (labeling, categorization, tagging)
- Transformation (format conversion, data restructuring)

**You can add more!** When you repeatedly create prompts for a new use case, ask Claude to add it to the taxonomy. This builds a customized library over time.

### Claude-Specific Optimization

The skill incorporates Claude-specific best practices:
- XML tag structure (Claude was trained on this)
- Chain-of-thought with `<thinking>` tags
- Specific over vague language
- Affirmative instructions (what TO do, not what NOT to do)
- Examples and input/output pairs
- Proper prompt ordering for long contexts

### Automated Analysis

The included Python script automatically checks for:
- Vague terms ("concise", "good", "detailed")
- Missing components (role, examples, format)
- Structure issues
- Negative instruction overuse
- Best practice gaps

## Example Workflow

Let's say you want to create a prompt for generating API documentation:

**Quick Mode (Default):**
```
You: "Create a prompt for API documentation"
Claude: "I'll create an API documentation prompt. Should this be for REST APIs, GraphQL, or general?"
You: "REST"
Claude: [creates complete prompt immediately]

"I've focused on REST endpoints with JSON. Includes request/response examples and error codes."

You: "Add authentication examples"
Claude: [updates prompt with auth examples]

You: "Perfect!"
```

**Guided Mode:**
```
You: "Help me think through a prompt for API documentation"
Claude: "Let's build this together. First, what type of APIs - REST, GraphQL, or something else?"
You: "REST APIs"
Claude: "Got it. Should the docs target internal developers or external API consumers?"
You: "External"
Claude: "Great. One more - do you need code examples in multiple languages, or just curl/HTTP?"
You: "Multiple languages"
Claude: [creates comprehensive prompt]

"Here's your prompt with multi-language examples and external-facing clarity."
```

**Expert Mode:**
```
You: "Quick prompt for REST API docs"
Claude: [creates complete prompt immediately]

"Assumed REST with JSON, multiple language examples, external audience. Ready to iterate!"
```

All three modes support the same natural iteration afterward!

## Conversational Features

### Natural References
The skill understands conversational references:
- "that prompt" → most recent prompt discussed
- "the last one" → most recent prompt
- "the prompt above" → prompt earlier in conversation
- "my prompt" → most recent user-provided prompt
- "your prompt" / "the one you created" → most recent Claude-generated prompt
- "it" → current working prompt

### Iterative Commands
Refine prompts naturally without re-pasting:
- **Length**: "make it shorter", "expand it", "add more detail"
- **Content**: "add examples", "include error handling", "focus on X"
- **Style**: "make it more formal", "simplify the language", "be more specific"
- **Structure**: "add XML tags", "reorganize this", "break it into sections"
- **Testing**: "test that with [example]", "show me how it would work"

### Multiple Prompts in Conversation
If you're working with several prompts:
- Claude will ask which one you mean if ambiguous
- You can refer to them by type: "the analysis prompt", "the code review one"
- Or by recency: "the first prompt", "the second one"

## Tips for Best Results

### Choosing Your Mode

1. **Use Quick Mode when**: You have a clear goal and want to get started fast
   - "Create a prompt for X"
   - Works for 80% of use cases

2. **Use Guided Mode when**: You're exploring and not sure what you need
   - "Help me think through..."
   - "I'm not sure what structure would work"

3. **Use Expert Mode when**: You know exactly what you want
   - "Quick prompt for X"
   - "Just make me a prompt that does Y"

### Working Conversationally

1. **Be specific about your goal**: "Create a prompt for Python code review" beats "Create a prompt"

2. **Share examples when helpful**: "Like this: [example]" gives Claude instant context

3. **Iterate freely**: Don't worry about getting it perfect upfront
   - Create → Test → "Make it more specific" → "Add examples" → Perfect!

4. **Use natural language**: "Make it shorter" works just as well as "reduce the length to 200 words"

5. **Build your library**: When you create great prompts, ask Claude to add them to examples.md

6. **Reference naturally**: "That prompt", "the last one", "the API one" all work

7. **Trust the modes**: The skill adapts to your style automatically

### Getting Unstuck

- **Too many questions?** Say "just create it, I'll refine as we go"
- **Not enough questions?** Say "walk me through this" to trigger Guided mode
- **Wrong assumptions?** Just correct: "Actually make it focus on X instead"

## Advanced: Extending the Skill

### Adding New Prompt Categories

When you find yourself creating similar prompts repeatedly:

1. Tell Claude: "I keep creating prompts for [use case]. Can we add this as a new category?"
2. Claude will update taxonomy.md with:
   - Purpose and characteristics
   - Essential elements
   - Common patterns
   - Anti-patterns
3. Add examples to examples.md
4. Your library grows!

### Customizing Best Practices

If you discover techniques that work well for your use case:
1. Ask Claude to add them to best-practices.md
2. Include examples and explanations
3. Your team can benefit from shared knowledge

## Troubleshooting

**"Claude isn't using the skill"**
- Make sure you're in a project where the skill is installed
- Use trigger phrases like "review this prompt" or "help me create a prompt"
- The skill triggers when you mention prompts or prompt engineering

**"The analysis seems generic"**
- Provide more context about what you're trying to achieve
- Share examples of desired output
- Specify your constraints and requirements

**"I want different suggestions"**
- Tell Claude what you'd prefer differently
- The skill is designed to iterate based on your feedback
- Don't hesitate to disagree and ask for alternatives

**"Claude didn't find my prompt from earlier"**
- Try being more specific: "the prompt about code review" instead of just "that prompt"
- If multiple prompts exist in the conversation, Claude will ask which one
- You can also just paste it again if the conversation is very long

**"Can I work on multiple prompts at once?"**
- Yes! Just be specific about which one: "audit the code review prompt" or "improve the first prompt"
- Claude will ask for clarification if there's ambiguity

## Getting Help

Within your project, you can:
- Ask Claude to explain any technique from best-practices.md
- Request to see examples from a specific category
- Ask "What prompt type should I use for [task]?"
- Request the skill to teach you about any technique

## What Makes This Special

Unlike generic prompt engineering advice, this skill:

1. **Adaptive Modes**: Automatically adjusts to your style (Quick/Guided/Expert)
2. **No Question Overload**: Max 1-2 questions per turn, never overwhelming lists
3. **Smart Assumptions**: Makes reasonable guesses and lets you correct
4. **Conversational Flow**: Works naturally with chat context - no re-pasting
5. **Claude-Optimized**: Uses Claude-specific techniques (XML, thinking tags, etc.)
6. **Immediate Generation**: Create mode generates prompts right away, iterate as needed
7. **Automated Analysis**: Script checks for common issues automatically
8. **Structured Workflow**: Clear process for audit/create/improve
9. **Extensible**: Grows with your needs via taxonomy
10. **Educational**: Explains WHY changes work (when you want details)

The conversation modes are key - whether you want to move fast (Quick), explore thoughtfully (Guided), or skip questions entirely (Expert), the skill adapts to you.

## Next Steps

1. Install the skill in your project
2. Try auditing an existing prompt you use
3. Create a new prompt for something you do often
4. Add your successful prompts to examples.md
5. Build your custom library over time

Enjoy better prompts and better results! 🚀
