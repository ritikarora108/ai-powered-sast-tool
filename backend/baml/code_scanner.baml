client OpenAI {
  provider openai
  api_key env("OPENAI_API_KEY")
}

prompt code_scanner {
  client OpenAI
  model "gpt-4-turbo"
  max_tokens 4000
  temperature 0.0

  inputs {
    code string
    language string
    filepath string
    vulnerability_types string
  }

  output CodeScanResult

  prompt ```
  You are a security expert performing an automated code scan for OWASP Top 10 vulnerabilities.
  Your task is to identify potential security vulnerabilities in the provided code.
  
  Please analyze the following code for these specific vulnerabilities: {{vulnerability_types}}
  
  Code language: {{language}}
  File path: {{filepath}}
  
  CODE:
  ```
  {{code}}
  ```
  
  Your task:
  1. Thoroughly analyze the provided code for security vulnerabilities.
  2. Focus on OWASP Top 10 vulnerabilities, especially the ones specified.
  3. For each vulnerability you find, provide:
     - Vulnerability type (from the OWASP Top 10)
     - Location (line numbers where the vulnerability exists)
     - Severity (Critical, High, Medium, Low)
     - Description of the vulnerability
     - A suggested remediation
  
  Provide output in a structured format that can be parsed programmatically.
  Only report actual vulnerabilities, not just code quality issues or stylistic improvements.
  If no vulnerabilities are found, return an empty list.
  ```
}

struct Vulnerability {
  vulnerability_type string
  line_start integer
  line_end integer
  severity string
  description string
  remediation string
  code_snippet string
}

struct CodeScanResult {
  vulnerabilities Vulnerability[]
} 