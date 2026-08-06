{
  description = "Backend Portal flake configurations";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  inputs.devenv.url = "github:cachix/devenv";
  inputs.devenv.inputs.nixpkgs.follows = "nixpkgs";

  outputs =
    inputs@{ flake-parts, nixpkgs, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.devenv.flakeModule ];
      systems = nixpkgs.lib.systems.flakeExposed;
      perSystem = _args: {
        devenv.shells.default =
          { config, pkgs, ... }:
          let
            appName = "web";
          in
          {
            env.DBNAME = "backend_portal";
            env.DBUSER = "root";
            env.DBPASSWORD = "1550";
            env.HOSTNAME = "localhost";
            env.GOPRIVATE = "github.com/paper-indonesia";

            packages = [
                pkgs.gopls # go language server protocol
                pkgs.go-outline # go outline
                pkgs.osv-scanner # go security scanner
                pkgs.gotools # go tools
                pkgs.gosec # go security scanner
                pkgs.gofumpt # format
                pkgs.golint # linting
                pkgs.gocyclo # calculate cyclomatic complexity
                
            ];

            # see full options: https://devenv.sh/supported-languages/go/
            languages.go.enable = true;
            languages.go.package = pkgs.go;

            # see full options: https://devenv.sh/supported-services/mysql/
            services.mysql.enable = true;
            services.mysql.package = pkgs.mysql80;
            services.mysql.ensureUsers = [
            {
                name = "myusername";
                password = "mypassword";
                ensurePermissions =
                {
                    "database.*" = "ALL PRIVILEGES";
                    "*.*" = "ALL PRIVILEGES";
                };
            }];
            services.mysql.initialDatabases = [ { name = "backend_portal"; } ];

            # see full options: https://devenv.sh/supported-services/rabbitmq/
            services.rabbitmq.enable = true;
            services.rabbitmq.listenAddress = "127.0.0.1";
            services.rabbitmq.managementPlugin.port = 15672;
            services.rabbitmq.managementPlugin.enable = true;
            services.rabbitmq.port = 5672;

            # see full options: https://devenv.sh/supported-services/redis/
            services.redis.enable = true;
            services.redis.bind = "127.0.0.1";
            services.redis.port = 6379;

            scripts.up.exec = # bash
              ''
                devenv up
              '';

          };
      };
    };
}