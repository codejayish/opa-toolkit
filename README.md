# OPA Toolkit

A unified Go library that simplifies policy development with the Open Policy Agent (OPA) by combining linting, formatting, testing, and benchmarking into one interface.

---

## Features

-  Linting with [Regal](https://github.com/StyraInc/regal)
-  Formatting using [OPA Formatter](https://www.openpolicyagent.org/docs/latest/tools/#format)
-  Testing using `opa test` with JSON and coverage support
-  Benchmarking using `opa bench`
-  Simple and unified interface for use in CI/CD or local development

---

## Installation

```bash
go get github.com/yourusername/opa-toolkit
