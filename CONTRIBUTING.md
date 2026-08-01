# Contributing to mapper

Thank you for your interest in contributing to `mapper`! We welcome bug reports, feature requests, documentation improvements, and code contributions.

Please follow these guidelines to make the contribution process smooth and efficient for everyone.

---

## Getting Started

### Prerequisites
* **Go**: Make sure you have Go installed (version `1.25.x` or later is recommended).
* **Git**: Needed to clone the repository and manage branches.

### Setup Instructions
1. Clone the repository:
   ```bash
   git clone https://github.com/audunhov/mapper.git
   cd mapper
   ```
2. Download dependencies:
   ```bash
   go mod download
   ```

---

## Development Workflow

### 1. Creating a Branch
Always create a descriptive branch for your changes:
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b bugfix/your-bug-description
```

### 2. Code Guidelines & Standards
* **Formatting**: Ensure your code is properly formatted before committing:
  ```bash
  go fmt ./...
  ```
* **Linting**: We use `golangci-lint` to maintain code quality. Run it locally or verify that the GitHub Actions lint checks pass on your PR.
* **Documentation**: If you introduce a new feature, update the corresponding documentation files (such as [`examples.md`](examples.md) or [`README.md`](README.md)) to reflect the new functionality.

### 3. Writing and Running Tests
Every new feature or bug fix should be accompanied by appropriate unit tests.
* To run the test suite locally:
  ```bash
  go test -v -race ./...
  ```
* Ensure all existing and new tests pass cleanly before submitting your PR.

---

## Submitting a Pull Request (PR)

1. Commit your changes with a clear, concise commit message:
   ```bash
   git commit -m "Support pointer fields in struct conversion"
   ```
2. Push your branch to GitHub:
   ```bash
   git push origin feature/your-feature-name
   ```
3. Open a Pull Request against the `main` branch.
4. Fill out the Pull Request template completely so reviewers understand what changes you've made and why.

---

## Code of Conduct

We expect all contributors to adhere to standard respectful open-source collaboration guidelines. Be kind, constructive, and helpful!
