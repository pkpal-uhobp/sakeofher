param(
    [ValidateSet("Preview", "Import", "Fix")]
    [string]$Mode = "Preview",

    [int64]$TariffID = 1,

    [string]$DbHost = $(if ($env:POSTGRES_HOST) { $env:POSTGRES_HOST } else { "127.0.0.1" }),
    [int]$DbPort = $(if ($env:POSTGRES_PORT) { [int]$env:POSTGRES_PORT } else { 5433 }),
    [string]$DbName = $(if ($env:POSTGRES_DB) { $env:POSTGRES_DB } elseif ($env:POSTGRES_DATABASE) { $env:POSTGRES_DATABASE } else { "sakeofher" }),
    [string]$DbUser = $(if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "postgres" }),
    [string]$DbPassword = $(if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { "postgres" }),

    # Optional manual override if auto-detect fails:
    # .\scripts\import-xui-users.ps1 -Mode Preview -DockerContainer your_postgres_container
    [string]$DockerContainer = "",

    # Optional manual override for docker compose service:
    # .\scripts\import-xui-users.ps1 -Mode Preview -ComposeService your_postgres_service
    [string]$ComposeService = "",

    [switch]$Yes
)

$ErrorActionPreference = "Stop"

$SqlDir = Join-Path $PSScriptRoot "sql"

switch ($Mode) {
    "Preview" { $SqlFile = Join-Path $SqlDir "preview_xui_import.sql" }
    "Import"  { $SqlFile = Join-Path $SqlDir "import_xui_to_sakeofher.sql" }
    "Fix"     { $SqlFile = Join-Path $SqlDir "fix_xui_import.sql" }
}

if (-not (Test-Path $SqlFile)) {
    throw "SQL file not found: $SqlFile"
}

Write-Host ""
Write-Host "Mode:      $Mode"
Write-Host "TariffID:  $TariffID"
Write-Host "Database:  ${DbUser}@${DbHost}:${DbPort}/${DbName}"
Write-Host "SQL file:  $SqlFile"
Write-Host ""

if ($Mode -ne "Preview" -and -not $Yes) {
    Write-Host "This will MODIFY the database."
    Write-Host "Type YES to continue:"
    $answer = Read-Host
    if ($answer -ne "YES") {
        Write-Host "Cancelled."
        exit 0
    }
}

function Invoke-LocalPsql {
    param([string]$File)

    Write-Host "Using local psql.exe"

    $old = $env:PGPASSWORD
    $env:PGPASSWORD = $DbPassword

    try {
        & psql `
            -h $DbHost `
            -p $DbPort `
            -U $DbUser `
            -d $DbName `
            -v ON_ERROR_STOP=1 `
            -v target_tariff_id=$TariffID `
            -f $File

        if ($LASTEXITCODE -ne 0) {
            throw "psql exited with code $LASTEXITCODE"
        }
    }
    finally {
        $env:PGPASSWORD = $old
    }
}

function Invoke-ComposePsql {
    param(
        [string]$File,
        [string]$Service
    )

    Write-Host "Using docker compose service: $Service"

    $sql = Get-Content -Raw -Encoding UTF8 $File

    $sql | docker compose exec `
        -T `
        -e PGPASSWORD=$DbPassword `
        $Service `
        psql `
            -U $DbUser `
            -d $DbName `
            -v ON_ERROR_STOP=1 `
            -v target_tariff_id=$TariffID

    if ($LASTEXITCODE -ne 0) {
        throw "docker compose psql exited with code $LASTEXITCODE"
    }
}

function Invoke-DockerContainerPsql {
    param(
        [string]$File,
        [string]$Container
    )

    Write-Host "Using docker container: $Container"

    $sql = Get-Content -Raw -Encoding UTF8 $File

    $sql | docker exec `
        -i `
        -e PGPASSWORD=$DbPassword `
        $Container `
        psql `
            -U $DbUser `
            -d $DbName `
            -v ON_ERROR_STOP=1 `
            -v target_tariff_id=$TariffID

    if ($LASTEXITCODE -ne 0) {
        throw "docker exec psql exited with code $LASTEXITCODE"
    }
}

function Find-ComposePostgresService {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        return $null
    }

    try {
        $services = @(docker compose ps --services 2>$null)
    } catch {
        return $null
    }

    if (-not $services -or $services.Count -eq 0) {
        return $null
    }

    foreach ($candidate in @("postgres", "db", "database", "postgresql", "pg")) {
        if ($services -contains $candidate) {
            return $candidate
        }
    }

    foreach ($service in $services) {
        if ($service -match "postgres|postgre|pg|db") {
            return $service
        }
    }

    return $null
}

function Find-PostgresContainer {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        return $null
    }

    $lines = @()
    try {
        $lines = @(docker ps --format "{{.Names}}`t{{.Image}}" 2>$null)
    } catch {
        return $null
    }

    foreach ($line in $lines) {
        $parts = $line -split "`t"
        if ($parts.Count -lt 2) {
            continue
        }

        $name = $parts[0]
        $image = $parts[1]

        if ($name -match "postgres|postgre|pg|db" -or $image -match "postgres|postgis") {
            return $name
        }
    }

    return $null
}

if (Get-Command psql -ErrorAction SilentlyContinue) {
    Invoke-LocalPsql -File $SqlFile
}
elseif ($ComposeService -ne "") {
    Invoke-ComposePsql -File $SqlFile -Service $ComposeService
}
elseif ($DockerContainer -ne "") {
    Invoke-DockerContainerPsql -File $SqlFile -Container $DockerContainer
}
elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    $service = Find-ComposePostgresService
    if ($service) {
        Invoke-ComposePsql -File $SqlFile -Service $service
    }
    else {
        $container = Find-PostgresContainer
        if ($container) {
            Invoke-DockerContainerPsql -File $SqlFile -Container $container
        }
        else {
            Write-Host ""
            Write-Host "Could not auto-detect PostgreSQL."
            Write-Host ""
            Write-Host "Run this and send/see the names:"
            Write-Host "  docker ps --format `"table {{.Names}}\t{{.Image}}\t{{.Ports}}`""
            Write-Host ""
            Write-Host "Then run with the container name, for example:"
            Write-Host "  .\scripts\import-xui-users.ps1 -Mode Preview -TariffID $TariffID -DockerContainer YOUR_CONTAINER_NAME"
            Write-Host ""
            throw "PostgreSQL container/service was not found."
        }
    }
}
else {
    throw "Neither psql nor docker was found. Install PostgreSQL client or run via Docker."
}

Write-Host ""
Write-Host "Done."

if ($Mode -ne "Preview") {
    Write-Host ""
    Write-Host "Next:"
    Write-Host "  make run-worker"
    Write-Host ""
    Write-Host "Worker will reconcile subscriptions/users with Remnawave."
}
