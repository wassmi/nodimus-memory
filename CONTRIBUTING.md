# Contributing to Nodimus Memory

We welcome contributions to Nodimus Memory! To ensure a smooth and effective collaboration, please follow these guidelines.

## How to Contribute

1.  **Fork the repository:** Start by forking the `nodimus-memory` repository to your GitHub account.
2.  **Clone your fork:** Clone your forked repository to your local machine:
    ```bash
    git clone https://github.com/YOUR_USERNAME/nodimus-memory.git
    cd nodimus-memory
    ```
3.  **Create a new branch:** Create a new branch for your feature or bug fix. Use a descriptive name:
    ```bash
    git checkout -b feature/your-feature-name
    # or
    git checkout -b bugfix/issue-description
    ```
4.  **Make your changes:** Implement your changes, ensuring they adhere to the existing code style and conventions.
5.  **Write tests:** If you're adding new features or fixing bugs, please write appropriate unit and integration tests to cover your changes.
6.  **Run tests:** Before committing, ensure all tests pass:
    ```bash
    go test ./...
    ```
7.  **Format your code:** Ensure your code is properly formatted:
    ```bash
    go fmt ./...
    ```
8.  **Commit your changes:** Write clear and concise commit messages. Reference any relevant issue numbers.
    ```bash
    git commit -m "feat: Add new feature for X" # or "fix: Resolve issue Y"
    ```
9.  **Push to your fork:** Push your changes to your forked repository:
    ```bash
    git push origin feature/your-feature-name
    ```
10. **Create a Pull Request (PR):** Open a pull request from your branch to the `main` branch of the original `nodimus-memory` repository. Provide a detailed description of your changes and why they are necessary.

## Code Style and Conventions

*   Follow standard Go formatting (`go fmt`).
*   Write clear, self-documenting code.
*   Add comments where the code's intent is not immediately obvious.
*   Ensure your code is well-tested.

## Reporting Bugs

If you find a bug, please open an issue on the GitHub issue tracker. Provide as much detail as possible, including:

*   A clear and concise description of the bug.
*   Steps to reproduce the behavior.
*   Expected behavior.
*   Actual behavior.
*   Your operating system and Go version.

## Feature Requests

We welcome ideas for new features! Please open an issue on the GitHub issue tracker to propose new features. Describe the feature, its potential benefits, and any relevant use cases.

## License

By contributing to Nodimus Memory, you agree that your contributions will be licensed under the MIT License.