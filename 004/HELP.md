# Übung für Cloud Computing - 004

Stelle sicher, dass [Terraform installiert](https://developer.hashicorp.com/terraform/downloads) ist. Prüfe dies in deinem Terminal:
```bash
terraform -version
```

## Der Infrastruktur-Code
Erstelle in diesem Verzeichnis eine Datei namens `main.tf` mit folgendem Inhalt:

```hcl
terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.9"
    }
  }
}

resource "local_file" "hello" {
  filename = "${path.module}/cloud_intro.txt"
  content  = "Infrastruktur als Code funktioniert lokal und in der Cloud!"
}
```

## Der Lifecycle-Workflow (Schritt-für-Schritt)

Führe nun die folgenden Befehle nacheinander in deinem Terminal aus und beobachte, was in deinem Verzeichnis passiert:

### 1. Initialisierung
```bash
terraform init
```
*Was passiert?* Terraform analysiert den Code und lädt das benötigte Provider-Plugin (`hashicorp/local`) in einen versteckten `.terraform` Ordner herunter.

### 2. Der Dry-Run (Planung)
```bash
terraform plan
```
*Was passiert?* Terraform zeigt dir einen Ausführungsplan. Ein grünes `+` signalisiert, dass eine neue Datei erstellt wird. Es wird noch nichts an deinem System verändert!

### 3. Das Deployment (Ausführung)
```bash
terraform apply
```
*(Bestätige die Abfrage mit `yes`)*

*Was passiert?* Die Datei `cloud_intro.txt` wird auf deiner Festplatte erstellt.
**Wichtig:** Prüfe dein Verzeichnis! Erst jetzt hat Terraform die Datei `terraform.tfstate` generiert, um sich dieses Deployment zu merken.

### 4. Der Rückbau (Aufräumen)
```bash
terraform destroy
```
*(Bestätige die Abfrage mit `yes`)*

*Was passiert?* Terraform liest den State, erkennt die verwaltete Textdatei und löscht sie sauber von deinem System. Der Ausgangszustand ist wiederhergestellt.