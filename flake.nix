{
  description = "Local PostgreSQL learning environment";

  inputs = {
    # This tracks a moving branch. Keep the generated flake.lock committed
    # so the actual nixpkgs revision remains reproducible.
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      # This currently supports Linux only.
      systems = [ "x86_64-linux" "aarch64-linux" ];

      forAllSystems = f:
        nixpkgs.lib.genAttrs systems (system:
          f {
            pkgs = import nixpkgs { inherit system; };
          });
    in {
      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            postgresql
          ];

          shellHook = ''
            export PGDATA="$PWD/.pgdata"

            # PGHOST is an absolute path, so PostgreSQL uses a Unix socket
            # inside the project directory rather than a system socket.
            export PGHOST="$PWD/.pgsocket"

            export PGPORT="55432"
            export PGDATABASE="learning"
            export PGUSER="$USER"

            mkdir -p "$PGHOST"

            # For a local learning database, trust authentication is convenient.
            # However, because PostgreSQL listens on localhost by default,
            # local TCP clients can also connect without a password.
            if [ ! -f "$PGDATA/PG_VERSION" ]; then
              echo "Initializing PostgreSQL database..."
              initdb \
                --pgdata="$PGDATA" \
                --auth=trust \
                --no-locale \
                --encoding=UTF8
            fi

            # This checks whether a server using this PGDATA is already running.
            # The connection options are not needed for `status`; pg_ctl primarily
            # checks the cluster's PID file.
            if ! pg_ctl \
              --pgdata="$PGDATA" \
              --options="-k $PGHOST -p $PGPORT" \
              --log="$PGDATA/postgresql.log" \
              status >/dev/null 2>&1
            then
              echo "Starting PostgreSQL on port $PGPORT..."
              pg_ctl \
                --pgdata="$PGDATA" \
                --options="-k $PGHOST -p $PGPORT" \
                --log="$PGDATA/postgresql.log" \
                --wait \
                start
            fi

            # This suppresses both "database already exists" and genuine errors,
            # such as PostgreSQL failing to accept connections. Consider checking
            # for existence explicitly if startup failures should be visible.
            createdb "$PGDATABASE" 2>/dev/null || true

            cleanup_postgres() {
              echo "Stopping PostgreSQL..."
              pg_ctl \
                --pgdata="$PGDATA" \
                stop \
                --mode=fast \
                >/dev/null 2>&1 || true
            }

            # Important lifecycle issue: this trap stops the server whenever this
            # shell exits, even if the server was already running before the shell
            # started. Multiple shells in the same directory can also interfere:
            # one shell exiting will stop the server used by the other.
            trap cleanup_postgres EXIT

            echo "PostgreSQL is running."
            echo "  Database: $PGDATABASE"
            echo "  Port:     $PGPORT"
            echo "  Socket:   $PGHOST"
            echo "  Connect:  psql"
          '';
        };
      });
    };
}

