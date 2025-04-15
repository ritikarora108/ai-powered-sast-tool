client OpenAI {
  provider openai
  api_key env("OPENAI_API_KEY")
}

prompt code_scanner {
  client OpenAI
  model "gpt-4o"
  max_tokens 16000
  temperature 0.0

  inputs {
    code string
    language string
    filepath string
    vulnerability_types string
  }

  output CodeScanResult

  prompt ```
  You are an expert security engineer performing static application security testing (SAST) for OWASP Top 10 vulnerabilities.
  Your task is to identify potential security vulnerabilities in the provided code with high precision.
  
  Please analyze the following code for these specific vulnerabilities: {{vulnerability_types}}
  
  Code language: {{language}}
  File path: {{filepath}}
  
  CODE:
  ```
  {{code}}
  ```
  
  Your analysis should:
  1. Focus specifically on the OWASP Top 10 vulnerabilities requested in the vulnerability_types field.
  2. Be precise and thorough - report only actual vulnerabilities, not merely code smells or style issues.
  3. Use context clues from the code to determine if something is truly a vulnerability or just a potential concern.
  4. Consider language-specific security patterns and anti-patterns for {{language}}.
  
  For each vulnerability found, provide:
  - Vulnerability type: The specific OWASP category (e.g., "A1:2021 - Broken Access Control")
  - Location: Exact line numbers (start and end) where the vulnerability exists
  - Severity: Critical, High, Medium, or Low, based on potential impact
  - Description: Clear explanation of the vulnerability and why it exists
  - Remediation: Specific, actionable steps to fix the vulnerability
  - Code snippet: The exact vulnerable code
  
  Common vulnerability patterns to look for:
  - Injection vulnerabilities (SQL, NoSQL, OS command, etc.)
  - Broken authentication mechanisms
  - Sensitive data exposure
  - XML External Entities (XXE)
  - Broken access control
  - Security misconfiguration
  - Cross-Site Scripting (XSS)
  - Insecure deserialization
  - Using components with known vulnerabilities
  - Insufficient logging & monitoring
  
  If no vulnerabilities matching the requested types are found, return an empty vulnerabilities list.
  
  Provide output in a structured format that can be parsed programmatically.
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