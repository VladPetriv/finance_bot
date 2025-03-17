# Balance Operations

### Flow with balances
List of available balance operations:

1. Create a new balance
2. Get balance information
3. Update balance
4. Delete balance


### Create a New Balance

During this flow, the user will be asked to enter the balance name, amount, and select a currency.

```mermaid
graph TD
    A[User is asked to enter the balance name] -->|If name exists| B[User is asked to enter another name]
    B --> A
    A --> C[User is asked to enter balance amount]

    C -->|If incorrect format| D[User is asked to enter another amount]
    D --> C
    C --> E[User is asked to select the currency from the list]

    E -->|If entered manually| F[User is asked to select currency from provided list]
    F --> E
    E --> G[End of the flow]
```

### Get Balance Information

During this flow, the user will be asked to select a month and a balance to view its information.

```mermaid
graph TD
    A[User is asked to select the month] --> B[User is asked to select their balance from the list]
    B --> C[End of the flow]
```

### Update Balance

During this flow, the user will be asked to select a balance and update its details.

```mermaid
graph TD
    A[User selects the balance to update] --> B[User enters the updated balance name or '-' to keep previous]
    B -->|If name exists| C[User is asked to enter another name]
    C --> B
    B --> D[User enters the updated balance amount or '-' to keep previous]

    D -->|If incorrect format| E[User is asked to enter amount in correct format]
    E --> D
    D --> F[User selects the currency from the list]

    F -->|If entered manually| G[User is asked to select currency from provided list]
    G --> F
    F --> H[End of the flow]
```


### Delete Balance

During this flow, the user will be asked to select a balance and confirm the deletion.

```mermaid
graph TD
    A[User selects the balance to delete] --> B[User confirms the action]
    B -->|If not confirmed| C[End of the flow]
    B -->|If confirmed| D[Associated operations are deleted in the background]
    D --> E[End of the flow]
```
