# User Instruction Memory

This file records user instructions, preferences, and teachings for reference in future interactions.

## Format

### User Instruction Entry
User instruction entries should follow this format:

[User Instruction Summary]
- Date: [YYYY-MM-DD]
- Context: [Mentioned scenario or time]
- Instructions:
  - [Content of user teaching or instruction, described line by line]

### Project Knowledge Entry
Entries discovered by the Agent during task execution should follow this format:

[Project Knowledge Summary]
- Date: [YYYY-MM-DD]
- Context: Discovered by Agent while performing [specific task description]
- Category: [Operations & Deployment|Build Methods|Testing Methods|Troubleshooting & Debugging|Workflow & Collaboration|Environment Configuration]
- Instructions:
  - [Specific knowledge points, described line by line]

## Deduplication Strategy
- Before adding a new entry, check for similar or identical instructions.
- If a duplicate is found, skip the new entry or merge it with the existing one.
- When merging, update the context or date information.
- This helps avoid redundant entries and keeps the memory file tidy.

## Entries

[Continue Through Confirmed Implementation Plan]
- Date: 2026-09-03
- Context: While implementing an approved multi-task feature plan
- Instructions:
  - Continue through the confirmed task list without pausing at intermediate review points.
  - Pause only for blockers that require user clarification or additional authorization.

[User Workflow Instructions]
- Date: 2026-09-05
- Context: OpsLang development and CI workflow
- Instructions:
  - Do not use subagents for this project.
  - Wait two seconds between API interactions when continuing project work.
  - After development and verification, push the changes and monitor GitHub Actions results.
