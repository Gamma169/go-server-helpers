# go-server-helpers

A library of helper functions I have made that are usable in many go-based microservices

## Usage

### Adding Package

You can add the package to your project the standard way in go using `go get`

```Shell
go get github.com/Gamma169/go-server-helpers
```

Or specify a version the standard way:

```Shell
go get github.com/Gamma169/go-server-helpers@v0.2.0
```

### Using Package

This library is organized into packages.  In order to use a function impot the package directly in your code.

**NOTE:** The capitalization in `Gamma169` is important or else the library will not be picked up.
```Go
import (
	"github.com/Gamma169/go-server-helpers/environments"
	"github.com/Gamma169/go-server-helpers/db"
	"github.com/Gamma169/go-server-helpers/server"
)

environments.GetRequiredEnv("SOME_VAR")

db.CheckAndRetry(...)

server.PreProcessInput(...)
```

Note that if there are package collisions (or if the name `environments` is too long and clunky), you can always use the standard package renaming in Go.

```Go
import (
	envs "github.com/Gamma169/go-server-helpers/environments"
)

envs.GetRequiredEnv("SOME_VAR")
```

This is used occasionally in the package.

## Package Versions + Changes

**NOTE:**  This package is a work in progress and subject to change.

Before v1.0.0 will use minor version tags to mark if the package exposed functions have changed.

**Minor Version Update:** If library goes from `v0.1.x` => `v0.2.x`, it means some function signatures have changed and **MAY NOT BE COMPATIBLE WITH PREVIOUS MINOR VERSION**

**Patch Version Update:** If library goes from `v0.2.3` => `v0.2.4`, it means new functions were added or bugs were fixed.  Library **WILL** be backwards-compatible with same minor version.


## Development

Developing on the library uses standard go tooling

### Get Dependencies
Use standard go get tool.
```Shell
go get ./...
```

### Build
For checking code compiles, does not build anything (yet-- have plans to create example package with `main` function that will run as an example)

```Shell
go build ./...
```

### Test
Note that tests are in their own directory and package.  This deviates slightly from go conventions.  But I could not see any significant downside to doing this.  

The main downside is that since the tests are in their own package (which is often recommended by documentation) one cannot test internal, unexported functions.  But since the exported functions are the most important, those are the ones tested (there are also very few unexported functoins in this library).  If an interal function needs to be tested, we can create `<file>_internal_test.go` files in the correct package.  Until then, place any tests in the `tests` directory.

```Shell
go test -v ./tests
```

### Format
Use standard gofmt tool.  **This formatting is enforced by the automated tests and tests will fail if code is not formatted properly**
```Shell
gofmt -w .
```

### One-liner that does everything (and formats your code for you)

```Shell
gofmt -w . && go get ./... && go build ./... && go test -v ./tests
```
