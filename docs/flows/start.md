### Start Flow

During this flow, the user will be asked to enter the name, amount, and currency of the initial balance.

```mermaid
graph TD
    A[User clicks on Start button] --> B[User enters the name of the first balance]
    B --> C[User enters the amount of the first balance]
    C --> D[User enters the currency of the first balance]
    D --> E[End of the flow. Show main menu to the user]
