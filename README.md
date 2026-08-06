<!-- Improved compatibility of back to top link -->
<a id="readme-top"></a>


<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/paper-indonesia/pivot-backoffice">
    <img src="https://github.com/user-attachments/assets/4a1bd5b7-1867-4b00-bb50-414ca45c06d4" alt="Logo" width="150" height="150">
  </a>

<h3 align="center">Backend Portal</h3>

  <p align="center">
    Backend Portal is a service that manage the entire business logic of the application. It is responsible for managing the data and the business logic. It is the core of the application, and it is the only part that is aware of the entire system.
    <br />
    <a href="https://github.com/paper-indonesia/pivot-backoffice/wiki"><strong>Explore the Wiki »</strong></a>
    <br />
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#docker">Run with Docker</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#owners">Owners</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>

<!-- GETTING STARTED -->
## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

You can run this project using either Docker or Nix. Make sure you have installed the following tools on your local machine:
* Docker
* Docker Compose
* Nix 2.18.5 or higher
* Go 1.22 or higher

### Docker

You can run this project using docker-compose. Make sure you have installed Docker and Docker Compose on your local machine.

1. Clone the repo
   ```sh
   git clone git@github.com:paper-indonesia/backend-portal.git
    ```
2. Go to the project directory
3. Copy from example config and secret
   ```sh
   cp .example.config.yaml .config.yaml
   cp .example.secret.yaml .secret.yaml
    ```
4. Download the dependencies
   ```sh
   go mod download -x
   ```
5. Run the project using docker-compose
   ```sh
   docker-compose up -d
   ```
6. Run the http project using makefile
   ```sh
   make run-http
   ```
7. The project will be running on `localhost:3000`, you can access health check on `localhost:3000/api/v1/health-check`
8. You can access the API documentation on `localhost:3000/swagger/index.html`

### Nix
We already put Go programming language configuration in .envrc file. You can use direnv to load the configuration.
Make sure you have installed Nix on your local machine.
1. Clone the repo
   ```sh
   git clone git@github.com:paper-indonesia/backend-portal.git
    ```
2. Go to the project directory
3. Allow the .envrc file to be loaded
   ```sh
   direnv allow
    ```
   The Go programming language configuration will be loaded. You can check it by running this command
   ```sh
    go env
     ```
4. Copy from example config and secret
   ```sh
   cp .example.config.yaml .config.yaml
   cp .example.secret.yaml .secret.yaml
    ```
5. Download the dependencies
   ```sh
   go mod download -x
   ```
6. Run the dependencies such as MySQL, Redis, and RabbitMQ
   ```sh
   nix develop --impure -c up
    ```
7. Open new terminal tab and run the project
   ```sh
   make run-http
    ```
8. The project will be running on `localhost:3000`, you can access health check on `localhost:3000/api/v1/health-check`

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

You can access the API documentation on `localhost:3000/swagger/index.html` or you can ask the team for the Postman collection.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

For other squad members who want to contribute to this project, you can follow these steps:

1. Clone the repo
   ```sh
   git clone git@github.com:paper-indonesia/backend-portal.git
    ```
2. Create your feature branch
   ```sh
   git checkout -b feat/[your-name]/[feature-name]
    ```   
   Example:
   ```sh
    git checkout -b feat/ericsson/add-new-endpoint
     ```
   Please refer to semantic for the branch name. You can read more about it [here](https://www.conventionalcommits.org/en/v1.0.0/).
3. Please make sure you run the test before pushing your code
   ```sh
   make test
    ```
   or run this command if you don't want to run the integration test. Yet I recommend you to run the integration test.
   ```sh
    make test --short
     ```
4. If you already run the test and everything is fine, you can run this command to generate the swagger documentation
   ```sh
   make swagger
    ```
5. Commit your changes
    ```sh
    git commit -m 'feat: Add some feature'
     ```
6. Push to the branch

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- REPOSITORY OWNERS -->
## Owners

This repository is maintained by the following squad members (backend team):
* [Ericsson Budhilaw](https://github.com/ebudhilaw)
* [Widya Ade Bagus](https://github.com/widyadebagus)
* [Muhammad Ariyanto Wijaya](https://github.com/jayharsya)
* [Rizaldy Septa Amanda](https://github.com/rizaldysepta-paper)
* [Achmad Ali B](https://github.com/achmad-alib)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

Please feel free to contact us if you have any questions or suggestions. You can reach us at:
* Slack: #payment-gateway-tech
* Or DM us directly

<p align="right">(<a href="#readme-top">back to top</a>)</p>