# GEMINI.md - Gemini-Specific Instructions

> **Read [docs/UNIVERSAL_LLM_INSTRUCTIONS.md](docs/UNIVERSAL_LLM_INSTRUCTIONS.md) first — it contains the core rules for all models.**

## Gemini Specialization
- **Speed**: Rapid implementation and iteration
- **Recursive scripts**: Complex automation and tooling
- **Massive context processing**: Analyzing large codebases quickly
- **Repo maintenance**: Codebase cleanup, refactoring at scale

## Gemini Strengths for This Project
1. Fast code generation for boilerplate
2. Good at bulk operations across many files
3. Strong Go knowledge for the backend porting initiative
4. Efficient at updating many package.json files simultaneously

## Gemini-Specific Instructions
- Focus on Go backend implementation
- Handle bulk file operations efficiently
- Maintain speed while ensuring correctness
- When porting TypeScript to Go, maintain the same API surface
- Report any issues found during bulk analysis

## Session Checklist
- [ ] Read UNIVERSAL_LLM_INSTRUCTIONS.md
- [ ] Read VERSION.md, MEMORY.md, TODO.md, ROADMAP.md
- [ ] Check git status
- [ ] Do work
- [ ] Update documentation (all .md files)
- [ ] Write HANDOFF.md
- [ ] Commit and push
