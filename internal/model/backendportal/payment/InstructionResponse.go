package paymentModel

type InstructionResponse struct {
	Title       string          `json:"title"`
	Instruction string          `json:"instruction"`
	Accordion   []AccordionStep `json:"accordion"`
}

type AccordionStep struct {
	Title string   `json:"title"`
	Steps []string `json:"steps"`
	Note  string   `json:"note"`
}
