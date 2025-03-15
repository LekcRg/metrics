# Exit on any error
set -e

# Function to build binaries
build_binaries() {
  echo "Building server and agent binaries..."
  go build -o ./cmd/server/server ./cmd/server/*.go && go build -o ./cmd/agent/agent ./cmd/agent/*.go
}

# Function to run statictest
run_statictest() {
  echo "Running statictest..."
  go1.22.12 vet -vettool=$(which statictest) ./...
}

# Function to run iteration 1
run_iteration1() {
  echo "Running Iteration 1 tests..."
  metricstest -test.v -test.run=^TestIteration1$ -binary-path=cmd/server/server
}

# Function to run iteration 2
run_iteration2() {
  echo "Running Iteration 2A tests..."
  metricstest -test.v -test.run=^TestIteration2A$ -source-path=. -agent-binary-path=cmd/agent/agent

  echo "Running Iteration 2B tests..."
  metricstest -test.v -test.run=^TestIteration2B$ -source-path=. -agent-binary-path=cmd/agent/agent
}

# Function to run iteration 3
run_iteration3() {
  echo "Running Iteration 3A tests..."
  metricstest -test.v -test.run=^TestIteration3A$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server

  echo "Running Iteration 3B tests..."
  metricstest -test.v -test.run=^TestIteration3B$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server
}

# Function to run iteration 4
run_iteration4() {
  echo "Running Iteration 4 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration4$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 5
run_iteration5() {
  echo "Running Iteration 5 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration5$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 6
run_iteration6() {
  echo "Running Iteration 6 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration6$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 7
run_iteration7() {
  echo "Running Iteration 7 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration7$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 8
run_iteration8() {
  echo "Running Iteration 8 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration8$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 9
run_iteration9() {
  echo "Running Iteration 9 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration9$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -file-storage-path=$TEMP_FILE -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 10
run_iteration10() {
  echo "Running Iteration 10A tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration10A$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -database-dsn='postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable' -server-port=$SERVER_PORT -source-path=.

  echo "Running Iteration 10B tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration10B$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -database-dsn='postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable' -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 11
run_iteration11() {
  echo "Running Iteration 11 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration11$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -database-dsn='postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable' -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 12
run_iteration12() {
  echo "Running Iteration 12 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration12$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -database-dsn='postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable' -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 13
run_iteration13() {
  echo "Running Iteration 13 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration13$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -database-dsn='postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable' -server-port=$SERVER_PORT -source-path=.
}

# Function to run iteration 14
run_iteration14() {
  echo "Running Iteration 14 tests..."
  SERVER_PORT=$(random unused-port)
  ADDRESS="localhost:${SERVER_PORT}"
  TEMP_FILE=$(random tempfile)
  metricstest -test.v -test.run=^TestIteration14$ -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server -database-dsn='postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable' -key="${TEMP_FILE}" -server-port=$SERVER_PORT -source-path=.
}

# Function to run a range of tests
run_range() {
  local start=$1
  local end=$2

  # Always build binaries first
  build_binaries

  # Always run statictest
  run_statictest

  # Run each iteration in the range
  for ((i=start; i<=end; i++)); do
    if [[ $i -ge 1 && $i -le 14 ]]; then
      eval "run_iteration$i"
    else
      echo "Invalid iteration: $i (must be between 1 and 14)"
      exit 1
    fi
  done

  echo "Tests for iterations $start to $end completed successfully!"
}

# Function to run a specific iteration
run_specific() {
  local iteration=$1

  # Always build binaries first
  build_binaries

  # Always run statictest
  run_statictest

  # Run the specific iteration
  if [[ $iteration -ge 1 && $iteration -le 14 ]]; then
    eval "run_iteration$iteration"
    echo "Tests for iteration $iteration completed successfully!"
  else
    echo "Invalid iteration: $iteration (must be between 1 and 14)"
    exit 1
  fi
}

# Function to run all tests
run_all() {
  build_binaries
  run_statictest
  run_iteration1
  run_iteration2
  run_iteration3
  run_iteration4
  run_iteration5
  run_iteration6
  run_iteration7
  run_iteration8
  run_iteration9
  run_iteration10
  run_iteration11
  run_iteration12
  run_iteration13
  run_iteration14
  echo "All tests passed successfully!"
}

# Main logic to process arguments
if [[ $# -eq 0 ]]; then
  # No arguments, run all tests
  run_all
elif [[ $1 =~ ^[0-9]+$ ]]; then
  # Single number argument, run specific iteration
  run_specific $1
elif [[ $1 =~ ^([0-9]+)-([0-9]+)$ ]]; then
  # Range argument, run specified range
  start=${BASH_REMATCH[1]}
  end=${BASH_REMATCH[2]}
  if [[ $start -le $end ]]; then
    run_range $start $end
  else
    echo "Invalid range: $start-$end (start must be less than or equal to end)"
    exit 1
  fi
else
  echo "Usage: $0 [iteration_number | start-end]"
  echo "  - No arguments: Run all tests"
  echo "  - iteration_number: Run a specific iteration (1-14)"
  echo "  - start-end: Run a range of iterations (e.g., 2-5)"
  exit 1
fi
