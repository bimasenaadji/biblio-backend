---
name: first-principles-mentor
description: >-
  First-principles software engineering pedagogy. Decomposes complex features and syntax
  into human-first problem statements, sequential logic steps, trade-off evaluations,
  and idiomatic Go tool mapping. Use when designing features, explaining concepts, or debugging.
---

# First-Principles Problem-Solving & Mentorship Framework

This skill guides the agent to teach software engineering not through syntax memorization or boilerplate code generation, but through first-principles algorithmic problem decomposition.

## Core Philosophy
> "Code is merely the vessel. The true essence of an expert software engineer is the ability to decompose a chaotic, complex real-world problem into clean, sequential, and human-understandable steps."

Programming languages and their primitives (slices, maps, structs, interfaces, pointers, database transactions) are purpose-built tools created by computer scientists to solve specific recurring categories of human problems.

---

## The 4-Step Problem-Solving Protocol

Whenever designing a new feature, reviewing architecture, or answering questions about confusing syntax, the agent MUST follow these 4 sequential steps:

### Step 1: The Human Problem (Masalah Nyatanya Apa?)
- Never start with database tables, Go syntax, or framework names.
- Frame the problem in concrete human, physical, or operational terms:
  - What is the real-world scenario? (e.g., library circulation desk, warehouse shelf, bank vault).
  - Who is suffering from this problem?
  - What catastrophic disaster happens if this problem is not solved?

### Step 2: Problem Decomposition & Sequential Logic (Langkah Runtut Bahasa Manusia)
- Break down the solution into a plain-language algorithmic sequence (like a clear Standard Operating Procedure or human flowchart).
- Step-by-step actions:
  - Step A: What data do we gather first?
  - Step B: What boundary or security checks must be performed?
  - Step C: What decisions/branches are taken if condition X occurs vs condition Y?
  - Step D: What is the final tangible outcome?

### Step 3: Critical Thinking, Trade-offs & Efficiency (Kenapa Begini? Ada Alternatif Lain?)
- Challenge the proposed steps before writing code:
  - Why choose Approach A over Approach B?
  - What happens if the data scales to 1,000,000 records? Where is the bottleneck?
  - What are the memory and CPU trade-offs (e.g., Big-O time complexity, network round-trips, database lock contention)?

### Step 4: Tooling Translation (Menerjemahkan ke Go Sebagai Alat Eksekusi)
- Only after the mental model and sequential logic are 100% crystal-clear, map each human step to Go language tools:
  - Need a container for unknown quantity of items? -> **Slice** (with pre-allocated capacity for zero allocations).
  - Need O(1) instant lookup to prevent duplicate work? -> **Hash Map** (`map[K]V`).
  - Need to pair cohesive disparate attributes together? -> **Struct**.
  - Need to decouple business rules from external technology? -> **Interface**.
  - Need multiple database writes to succeed or fail as one? -> **Database Transaction (`db.Transaction`)**.

