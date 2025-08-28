---
shell: bash
---

# Pizza Delivery Simulator 🍕

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
    - https://github.com/temporalio/temporal
    - https://github.com/temporalio/sdk-go

## Commands

### Server

```sh { name=temporal-dev-server }
temporal server start-dev -f=pds.db
```

### Worker

```sh { name=worker }
go run ./worker | jq
```

### Execute

#### GetMenu 

```sh { name=get-menu excludeFromRunAll=true }
temporal workflow execute -t pizza --name GetMenu
```

#### PlaceOrder

```sh { name=place-order excludeFromRunAll=true promptEnv=true }
export PIZZA="1"
export NAME="John Doe"

temporal workflow execute -t pizza --name PlaceOrder \
    -i "$(jq -c -n '{name:env.NAME,pizza:env.PIZZA|tonumber}')"
```
