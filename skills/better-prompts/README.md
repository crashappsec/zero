<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Prompt Engineering Skill - Complete Package

Your conversational prompt engineering assistant is ready! Here's everything included.

---

## 📦 Core Files

### [prompt-engineering.skill](computer:///mnt/user-data/outputs/prompt-engineering.skill)
**The main skill file** - Upload this to your claude.ai project to activate the skill.

**What's inside:**
- SKILL.md with complete workflow
- Automated analysis script (analyze_prompt.py)
- Comprehensive references (taxonomy, best practices, examples)

---

## 📚 Documentation

### [prompt-engineering-guide.md](computer:///mnt/user-data/outputs/prompt-engineering-guide.md)
**User guide and getting started**

**Covers:**
- Installation instructions
- Conversation modes (Quick/Guided/Expert)
- Usage patterns
- Feature overview
- Tips for best results
- Troubleshooting

**Start here if you're new!**

---

### [example-conversations.md](computer:///mnt/user-data/outputs/example-conversations.md)
**Real conversations showing all three modes**

**Includes:**
- Quick Mode example (Python code review)
- Guided Mode example (customer support bot)
- Expert Mode example (meeting notes)
- Audit workflow example
- Improve workflow example

**Great for understanding how it actually works in practice.**

---

### [before-after-comparison.md](computer:///mnt/user-data/outputs/before-after-comparison.md)
**Shows the conversation flow improvements**

**Demonstrates:**
- Old way: question overload → New way: 1-2 questions max
- Old way: verbose analysis → New way: scannable feedback
- Old way: re-paste everything → New way: natural references
- Design principles applied

**Illustrates why the new approach is better.**

---

## 🎯 Key Features

### Three Conversation Modes

**🚀 Quick Mode (Default)**
- 0-1 questions
- Immediate generation
- Smart assumptions
- Fast iteration

**🤔 Guided Mode**
- 1-2 questions per turn
- Progressive understanding
- Thoughtful exploration
- Comprehensive results

**⚡ Expert Mode**
- Zero questions
- Instant generation
- For experienced users
- Minimal refinement

### Core Capabilities

1. **Create prompts** - From scratch with adaptive questioning
2. **Audit prompts** - Scannable feedback in 5 seconds
3. **Improve prompts** - Fix failures immediately
4. **Iterate naturally** - "make it shorter", "add examples", "more formal"

### Technical Excellence

- **Automated analysis** - Python script catches common issues
- **Claude-optimized** - XML tags, thinking tags, specific techniques
- **Extensible taxonomy** - 8 prompt types, add your own
- **Reference library** - Best practices, examples, patterns
- **Context-aware** - Works with conversation, no re-pasting

---

## 🚀 Quick Start

1. **Install**: Upload `prompt-engineering.skill` to your claude.ai project
2. **Try it**: "Create a prompt for [your task]"
3. **Iterate**: "Make it shorter", "Add examples", "More specific"
4. **Audit**: "Review this prompt: [paste]"
5. **Improve**: "This didn't work, fix it: [paste]"

---

## 💡 Usage Examples

### Creating (Quick Mode)
```
You: "Create a prompt for data analysis"
Claude: "What type of data - sales, user behavior, or financial?"
You: "Sales"
Claude: [creates prompt immediately with smart defaults]
```

### Creating (Expert Mode)
```
You: "Quick prompt for API documentation"
Claude: [creates complete prompt instantly]
```

### Auditing
```
You: "Audit this prompt: [paste]"
Claude: [Scannable feedback in 5 seconds]
"This is functional but could be more specific..."
```

### Improving
```
You: "This didn't work, it gave bullets but I wanted prose"
Claude: [Shows fixed version immediately]
"Added explicit prose format requirement..."
```

### Iterating
```
You: "Make it shorter"
Claude: [Updated version] "Condensed to 150 words."
You: "Add examples"
Claude: [Adds examples]
You: "Perfect!"
```

---

## 📖 What to Read

### If you're new:
1. Start with **prompt-engineering-guide.md**
2. Skim **example-conversations.md** to see it in action
3. Install and try it!

### If you want to understand the improvements:
1. Read **before-after-comparison.md**
2. See how conversation flow improved
3. Understand design principles

### If you want concrete examples:
1. Jump to **example-conversations.md**
2. See all three modes in action
3. Learn usage patterns

---

## 🎨 Conversation Mode Selection

The skill automatically detects which mode you want:

| Your Language | Mode Triggered |
|---------------|----------------|
| "Create a prompt for..." | Quick Mode |
| "Help me think through..." | Guided Mode |
| "Quick prompt for..." | Expert Mode |
| "Walk me through..." | Guided Mode |
| "I'm not sure..." | Guided Mode |
| "Just make me a prompt..." | Expert Mode |

You can also explicitly request: "Use guided mode" or "Keep it quick"

---

## 🔧 Technical Details

### What's in the Skill

**SKILL.md** (main workflow):
- Conversation mode detection
- Three workflows (audit/create/improve)
- Context-aware prompt location
- Minimal questioning approach
- Smart approval requirements

**scripts/analyze_prompt.py**:
- Detects vague language
- Checks for missing components
- Identifies best practice gaps
- Provides structured feedback

**references/taxonomy.md**:
- 8 extensible prompt categories
- Characteristics and patterns
- Essential elements
- Common anti-patterns

**references/best-practices.md**:
- Claude-specific techniques
- XML structure, thinking tags
- Common pitfalls
- Optimization strategies

**references/examples.md**:
- Ready-to-use prompts
- Organized by category
- Input/output pairs
- Real-world patterns

---

## 🎯 Design Principles

1. **Bias toward action** - Generate first, refine later
2. **Minimal friction** - Remove unnecessary gates
3. **Smart defaults** - Assume and let user correct
4. **Progressive disclosure** - Show what matters
5. **Conversational flow** - Natural dialogue
6. **Respect time** - Scannable feedback
7. **Support iteration** - Easy refinement

---

## ✨ What Makes This Special

Unlike generic prompt engineering tools:

✅ Adapts to your working style (3 modes)
✅ Never overwhelms with questions
✅ Generates immediately in most cases
✅ Works conversationally (no re-pasting)
✅ Claude-specific optimizations
✅ Automated analysis catches issues
✅ Extensible taxonomy grows with you
✅ Educational when you want details

---

## 🤝 Getting Help

Once installed, you can ask Claude:

- "Explain [technique] from best practices"
- "Show me examples of [prompt type]"
- "What prompt type should I use for [task]?"
- "Walk me through creating [prompt]"
- "Why did you suggest [change]?"

The skill is designed to teach as it works.

---

## 🔮 Next Steps

1. **Install the skill** in your claude.ai project
2. **Try Quick Mode** - "Create a prompt for [something you do often]"
3. **Test iteration** - "Make it shorter", "Add examples"
4. **Audit an existing prompt** you use regularly
5. **Try Guided Mode** - "Help me think through [complex requirement]"
6. **Build your library** - Add successful prompts to examples.md
7. **Extend taxonomy** - Add new categories as you discover patterns

---

## 📊 Comparison: Before vs After

| Metric | Before | After |
|--------|--------|-------|
| Questions per turn | 5-8+ | 0-2 max |
| Time to first prompt | 3-5 turns | Often 1 turn |
| Audit feedback length | 500+ words | 50-100 words |
| Re-pasting required | Every iteration | Never |
| Approval gates | Always | Only when appropriate |
| Mode options | One size fits all | 3 adaptive modes |
| User experience | Overwhelming | Conversational |

---

## 🎉 You're Ready!

You now have a conversational prompt engineering assistant that:
- Adapts to your style
- Keeps conversation flowing
- Never overwhelms with questions
- Generates quickly
- Iterates naturally
- Teaches as it works

Install the `.skill` file and start creating better prompts through natural conversation!

**Pro tip**: Start with Quick Mode for 80% of your needs. Try Guided Mode when you're exploring something new. Use Expert Mode when you want zero friction.

Happy prompting! 🚀
