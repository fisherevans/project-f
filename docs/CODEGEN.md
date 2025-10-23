# Code Generation

This project uses `go generate` to automatically generate TypeScript definitions for the scripting API.

## TypeScript Definitions

The following files are **auto-generated** from Go source code:
- `assets/scripts/global.d.ts` - TypeScript type definitions
- `assets/scripts/template.js` - JSDoc-annotated template for new scripts

### Regenerating

```bash
go generate ./internal/game/events
```

### Using in Your Scripts

All scripts must export a `handler` object:

```javascript
// @ts-check

/** @type {import('./global').EntityHandler} */
const handler = {
    Init(self, world, state) {
        return {
            state: state || {},
        };
    },

    OnInteract(self, world, state, event) {
        return {
            state: state,
            effects: [
                {
                    mutateEntity: {
                        entityId: self.Id(),
                        mode: "active"  // ✅ Autocomplete works here!
                    }
                }
            ]
        };
    }
};
```

### Benefits

- ✅ **One type annotation** for the entire handler
- ✅ **Full autocomplete** for all fields (parameters, effects, events)
- ✅ **Type checking** catches typos (e.g., `cat: "hey"` will error)
- ✅ **Inline documentation** in your IDE
- ✅ **Clean syntax** - no JSDoc needed on every function

### What Gets Generated

1. **Event Types** - Extracted from `internal/game/events/handler.go`
   - Reads `registerGojaEventHandler[T]("FunctionName")` calls
   - Generates TypeScript interfaces for each event type
   - Creates handler function signatures

2. **Effect Types** - Extracted from `internal/game/events/effect.go`
   - Generates interfaces for all `Effect*` structs
   - Includes helper types (e.g., `DynamicAnimationReference`, `LightConfig`)
   - Creates the main `Effect` union type

### Adding a New Event

1. Define the event struct in `handler.go`:
   ```go
   type EventPlayerDeath struct {
       PlayerId string
       Cause    string
   }
   ```

2. Register it in `init()`:
   ```go
   func init() {
       registerGojaEventHandler[EventPlayerDeath]("OnPlayerDeath")
   }
   ```

3. Run `go generate ./internal/game/events`

4. TypeScript definition is automatically created:
   ```typescript
   interface EventPlayerDeath {
       playerId: string;
       cause: string;
   }
   
   declare function OnPlayerDeath(
       self: EntityContext,
       world: WorldStateReader,
       state: any,
       event: EventPlayerDeath
   ): HandlerOutput | null;
   ```

### Type Mapping

| Go Type | TypeScript Type |
|---------|----------------|
| `string` | `string` |
| `int`, `float64` | `number` |
| `bool` | `boolean` |
| `*T` | `T \| undefined` |
| `[]T` | `T[]` |
| `map[K]V` | `Record<K, V>` |

### Implementation

The generator is implemented in `internal/game/events/generate_types.go` and uses Go's AST parser to extract type information directly from source code.
