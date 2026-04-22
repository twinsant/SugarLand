# SugarLand Metaverse No. 25
AI powered Sugarscape

## Project Summary

SugarLand is an AI-powered reimplementation of the classic [Sugarscape](https://en.wikipedia.org/wiki/Sugarscape) agent-based model (ABM), originally conceived by Epstein & Axtell in *Growing Artificial Societies* (1996). The project replaces traditional hard-coded agents with **AI Agents**, enabling richer emergent social behavior in a simulated economy.

### Key Features

| Feature | Description |
|---|---|
| **AI Agent Simulation** | Agents are driven by AI instead of simple rule scripts, allowing more realistic decision-making |
| **Dual-Resource Economy** | Sugar and Spice resources with seasonal cycles, pollution, and regeneration mechanics |
| **Social Behaviors** | Barter trading (Bull/Bear strategies), reproduction with genetic inheritance, and wealth transfer |
| **50×50 Torus Grid** | Wrap-around topology eliminates edge bias; dual-peak resource topography drives migration |
| **Population Dynamics** | Birth, aging, death, natural selection pressure, and emergent wealth inequality (Gini coefficient) |
| **Real-time Monitoring** | Scoreboard with population stats, attribute distributions, wealth Gini, and performance metrics |

### Technical Architecture

- **Cellspace**: 50×50 grid with Torus topology, dual-peak resource landscape, seasonal growth rates, and pollution accumulation.
- **Citizens**: Heterogeneous agents with genetically determined vision (*v* ∈ [1,6]), metabolism (*m* ∈ [1,4]), max age ∈ [60,100], and initial wealth ∈ [5,25].
- **Ruleset Engine**: Strict G→M→R execution order (Grow → Move → Replace) with randomized agent scheduling to eliminate first-mover bias.
- **Advanced Economics**: Marginal-value-based barter with Bull (high-frequency) and Bear (safety-margin) trading strategies.

### Project Goal

The practical significance of this project is to provide an accessible platform for exploring questions about wealth inequality, market dynamics, and social structure through AI-driven simulation — questions such as *why are the rich rich and the poor poor?*, *what can free markets solve and what can't they?*, and *how do inheritance and initial endowments shape social stratification?*

> For the full technical specification, see [SPEC.md](SPEC.md).

## What is Sugarscape?

Sugarscape is a simulation game designed by Thomas Schelling, an economist at the University of Maryland, in 1969. However, its more significant purpose is to conduct a complex economic experiment to study private property, consumption, and other aspects of human society.

> Horizontal Inequality

Understanding the Sugarscape experiment allows us to think more deeply about capitalism, private property, the free market, trade, finance, and all the issues you care about. To put it simply, through the Sugarscape experiment, you can observe a [...]

* Why are you poor?
* Why are they rich? Why do I have to spend my whole life to achieve the standard of living that a "rich second generation" is born with?
* Is it really because you don't work hard enough that you can't afford a 79-yuan eyebrow pencil?
* What problems can the free market solve, and what can't it? How can the state provide value in the field of public goods?
* Why should we study Marxism? On the eve of AI potentially dominating human society, why is it important to understand the need for a market economy while adhering to public ownership of the means of production and socialism? [...] 

The Sugarscape experiment is an excellent social experiment game created by international scholars moving away from metaphysical traditional economics. It also provides a great simulation environment for everyone involved in Web3 who hopes to use blockchain technology to transform production relations.

## Game Design of the Sugarscape Experiment

Please refer to ["The Origin of Wealth"](https://book.douban.com/subject/34834004/). I won't go into detail here.

## AI Sugarscape Experiment

Based on the design of the Sugarscape experiment, we found that the agents are very suitable for simulation using AI Agents. This is the original intention of this project. "Ant" (the author) believes the practical significance of this project far exceeds the "Stanford Small Town" [...]

We welcome everyone interested in economics to participate, to learn and think about more things together.

This project is divided into two repositories (domestic and international) for convenience:

* Main GitHub Repo: https://github.com/twinsant/sugarland/
* (Deprecated (●`ε´●) Gitee Mirror

# Refs

* [NotebookLM Resources](https://notebooklm.google.com/notebook/95e9cbad-8a41-4be3-945b-06ef154a615e)
  * [Sugarscape (Wikipedia)](https://en.wikipedia.org/wiki/Sugarscape)
  * [The Sugarscape (SourceForge)](https://sugarscape.sourceforge.net/)
  * [Sugarscape (AgentsExampleZoo.jl)](https://juliadynamics.github.io/AgentsExampleZoo.jl/dev/examples/sugarscape/)
