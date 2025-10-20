package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Operation represent a financial operation.
type Operation struct {
	ID                    string `db:"id"`
	CategoryID            string `db:"category_id"`
	BalanceID             string `db:"balance_id"`
	BalanceSubscriptionID string `db:"balance_subscription_id"`
	ParentOperationID     string `db:"parent_operation_id"`

	Type         OperationType `db:"type"`
	Amount       string        `db:"amount"`
	Description  string        `db:"description"`
	ExchangeRate string        `db:"exchange_rate"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var typeShort = map[OperationType]string{
	OperationTypeIncoming:    "IN",
	OperationTypeSpending:    "OUT",
	OperationTypeTransfer:    "TRF",
	OperationTypeTransferIn:  "TRF-IN",
	OperationTypeTransferOut: "TRF-OUT",
}

// GetName returns a string representation of the operation.
func (o Operation) GetName() string {
	truncatedDescription := o.Description
	if len(o.Description) > 10 {
		truncatedDescription = o.Description[:10] + "..."
	}

	return fmt.Sprintf("%s, %s, %s, %s",
		o.CreatedAt.Format("02.01 15:04:05"),
		o.Amount,
		typeShort[o.Type],
		truncatedDescription,
	)
}

// GetID returns the ID of the operation.
func (o Operation) GetID() string {
	return o.ID
}

// GetDeletionMessage generates a confirmation message for operation deletion,
// including operation details and warnings about balance impacts based on the
// operation type (transfer, spending, or incoming).
func (o *Operation) GetDeletionMessage() string {
	var balanceImpactMsg string
	switch o.Type {
	case OperationTypeTransferIn, OperationTypeTransferOut:
		balanceImpactMsg = fmt.Sprintf(
			"⚠️ Warning: Deleting this transfer operation will:\n"+
				"• Increase balance of source account by %s\n"+
				"• Decrease balance of destination account by %s",
			o.Amount,
			o.Amount)
	case OperationTypeSpending, OperationTypeIncoming:
		var effect string
		if o.Type == "income" {
			effect = "decrease"
		} else {
			effect = "increase"
		}
		balanceImpactMsg = fmt.Sprintf(
			"⚠️ Warning: Deleting this %s operation will %s your balance by %s",
			o.Type,
			effect,
			o.Amount)
	}

	return fmt.Sprintf(
		"Are you sure you want to proceed with the selected operation?\n\n"+
			"Operation Details:\n"+
			"Type: %s\n"+
			"Amount: %s\n"+
			"Description: %s\n\n"+
			"%s\n\n"+
			"This action cannot be undone.",
		o.Type, o.Amount, o.Description,
		balanceImpactMsg,
	)
}

// GetDetails returns the operation details in string format.
func (o *Operation) GetDetails() string {
	return fmt.Sprintf(
		"Operation Details:\nType: %s\nAmount: %s\nDescription: %s",
		o.Type, o.Amount, o.Description,
	)
}

// OperationType represents the type of an operation, which can be either incoming, spending or transfer.
type OperationType string

const (
	// OperationTypeIncoming represents a incoming operation.
	OperationTypeIncoming OperationType = "incoming"
	// OperationTypeSpending represents a spending operation.
	OperationTypeSpending OperationType = "spending"
	// OperationTypeTransfer represents a transfer operation.
	OperationTypeTransfer OperationType = "transfer"
	// OperationTypeTransferIn represents a transfer_in operation.
	OperationTypeTransferIn OperationType = "transfer_in"
	// OperationTypeTransferOut represents a transfer_out operation.
	OperationTypeTransferOut OperationType = "transfer_out"
)

type createOperationPromptData struct {
	UserInput  string     `json:"user_input"`
	Categories []Category `json:"categories"`
}

// BuildCreateOperationFromTextPrompt builds a prompt for creating an operation based on provided categories and text from user.
func BuildCreateOperationFromTextPrompt(userInput string, categories []Category) (string, error) {
	basePromptTemplate := `You are a financial text parser. Extract structured data from user input and match the correct category.
### **Instructions**:
- **Input:** JSON with a category list and a financial text entry.
- **Output:** A JSON with:
  - "amount": Extracted **numeric string** (e.g., "10.00", "500", "123.31").
  - "category_id": The **UUID** of the best-matching category or an **empty string** ("") if no category can be determined.
  - "description": Extracted **description** (e.g., "Salary", "Food").
  - "type": Could be incoming | spending based on user input
- **Rules**:
  - Amounts always follow this format: 100.12, 100, 123.31 (no commas).
  - Negative (-) = **expense**, Positive (+) = **income**.
  - Select the **most relevant category** based on the text.
  - **If no suitable category is found, return "category_id": ""** (do not invent a category).
  - Return **only JSON**, nothing else.

### **User Input & Categories**:
%s

### **Expected Output Format**:
{
  "amount": "10.38",
  "description": "Salary",
  "category_id": "",
  "type": "incoming"
}`

	encodedPromptData, err := json.Marshal(createOperationPromptData{
		UserInput:  userInput,
		Categories: categories,
	})
	if err != nil {
		return "", fmt.Errorf("marshal categories: %w", err)
	}

	return fmt.Sprintf(basePromptTemplate, string(encodedPromptData)), nil
}

// OperationData represents the data extracted from the prompt output.
type OperationData struct {
	Amount      string        `json:"amount"`
	Description string        `json:"description"`
	CategoryID  string        `json:"category_id"`
	Type        OperationType `json:"type"`
}

// OperationDataFromPromptOutput parses the output from the prompt and returns the OperationData.
func OperationDataFromPromptOutput(output string) (*OperationData, error) {
	output = strings.ReplaceAll(output, "```json", "")
	output = strings.ReplaceAll(output, "```", "")

	var data OperationData
	err := json.Unmarshal([]byte(output), &data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal operation data: %w", err)
	}

	return &data, nil
}
