package templates

import (
    "bytes"
    "text/template"
)

// ProcessTemplate procesa un template con los datos proporcionados
func ProcessTemplate(templateStr string, data interface{}) (string, error) {
    tmpl, err := template.New("email").Parse(templateStr)
    if err != nil {
        return "", err
    }

    var result bytes.Buffer
    if err := tmpl.Execute(&result, data); err != nil {
        return "", err
    }

    return result.String(), nil
}

// ValidateTemplate verifica si un template es válido
func ValidateTemplate(templateStr string) error {
    _, err := template.New("email").Parse(templateStr)
    return err
} 