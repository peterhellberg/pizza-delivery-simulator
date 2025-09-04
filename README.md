# Pizza Delivery Simulator 🍕

> [!Tip]
> This is a notebook that can be used with the [Runme CLI](https://docs.runme.dev/installation/cli) 📚
>
> _Installation of the Runme CLI:_
>   - **On macOS:** `brew install runme`
>   - **Using Go:** `go install github.com/stateful/runme@latest`
>   - **Binaries:** <https://github.com/stateful/runme/releases>

## Plan _(not all implemented yet)_

- Workflow: Customer places an order.
- Activities:
    - Bake pizza _(takes time, might fail/retry if the "oven breaks")_.
    - Dispatch delivery driver.
    - Driver reports progress via Temporal signals _(e.g., “Stuck in traffic”, “Almost there”)_.
- Features shown: Long-running workflow, async signals, retries, timers (simulate delays).

## Dependencies

- https://go.dev/
    - https://github.com/golang/go
- https://temporal.io/
    - https://temporal.io/setup/install-temporal-cli
    - https://github.com/temporalio/temporal
    - https://github.com/temporalio/sdk-go
- https://jqlang.org/

## Commands

### Execute

#### GetMenu 

```sh { name=get-menu excludeFromRunAll=true }
temporal workflow execute -t pizza --name GetMenu
```

#### PlaceOrder

```sh { name=place-order excludeFromRunAll=true promptEnv=true }
export PIZZA="5"
export NAME="Peter"

temporal workflow execute -t pizza --name PlaceOrder \
    -i "$(jq -c -n '{name:env.NAME,pizza:env.PIZZA|tonumber}')"
```

### Server

```sh { name=temporal-dev-server }
temporal server start-dev -f=pds.db
```

### Search attributes

```sh { name=temporal-create-search-attributes }
temporal operator search-attribute create --name DriverAssigned --type Bool
temporal operator search-attribute create --name OrderID --type Text
```

### Worker

```sh { name=worker }
go run ./worker | jq
```

### Install `temporal` and `jq` using [Homebrew](https://brew.sh/) 🍏

```sh { name=brew-install excludeFromRunAll=true }
# Install temporal and jq using brew
if ! [ -x "$(command -v brew)" ]; then
    echo 'Error: brew is not installed.' >&2
    exit 1
fi

brew install temporal jq
```
