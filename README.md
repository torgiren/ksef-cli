<div align="center">

<img src="https://www.podatki.gov.pl/assets/logoPodatki.svg" alt="podatki.gov.pl" height="40"/>

# ksef-cli

**Tekstowy klient wiersza poleceń dla Krajowego Systemu e-Faktur**

[![Build](https://github.com/torgiren/ksef-cli/actions/workflows/go.yml/badge.svg)](https://github.com/torgiren/ksef-cli/actions/workflows/go.yml)
[![Release](https://github.com/torgiren/ksef-cli/actions/workflows/release.yml/badge.svg)](https://github.com/torgiren/ksef-cli/actions/workflows/release.yml)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)
[![KSeF](https://img.shields.io/badge/KSeF-API_v2-013f71)](https://www.podatki.gov.pl/ksef/)

</div>

---

Nieoficjalny klient CLI dla [Krajowego Systemu e-Faktur (KSeF)](https://www.podatki.gov.pl/ksef/) — polskiego systemu faktur elektronicznych Ministerstwa Finansów.

## Co potrafi

| | Funkcja |
|--|---------|
| ✅ | **Logowanie** do KSeF przy użyciu NIP-u i tokenu KSeF |
| ✅ | **Zarządzanie profilami** — przechowywanie wielu konfiguracji (różne firmy/NIP-y) |
| ✅ | **Automatyczne odświeżanie tokenów** — tokeny są cachowane i odświeżane bez ponownego logowania |
| ✅ | **Listowanie faktur** — pobieranie listy faktur z KSeF (domyślnie ostatnie 3 miesiące) |
| ✅ | **Wiele formatów wyjścia** — tekst (domyślnie) lub JSON |
| ✅ | **Poziomy logowania** — od cichego do pełnego podglądu żądań API (`-v` / `-vv` / `-vvv`) |

## Czego jeszcze nie potrafi

| | Funkcja |
|--|---------|
| ❌ | **Pobieranie plików faktur** — dostępna jest tylko lista metadanych, nie da się pobrać faktury jako XML/PDF |
| ❌ | **Wysyłanie faktur** — brak możliwości przesłania nowej faktury do KSeF |
| ❌ | **Zarządzanie profilami z poziomu CLI** — `profile list`, `profile set` i `profile delete` są niezaimplementowane |
| ❌ | **Operacje wsadowe** |

## Instalacja

### Pobranie gotowego binarki

Pliki binarne dla Linux, macOS i Windows dostępne są w [Releases](../../releases).

### Budowanie ze źródeł

Wymagany Go 1.25+.

```bash
git clone https://github.com/torgiren/ksef-cli
cd ksef-cli
make build
```

Binarka `ksef-cli` pojawi się w bieżącym katalogu.

## Konfiguracja

Plik konfiguracyjny tworzony jest automatycznie przy pierwszym logowaniu:

- **Konfiguracja:** `~/.config/ksef-cli/config.yaml`
- **Cache tokenów:** `~/.cache/ksef-cli/profile_<nazwa>.json`

## Użycie

### Logowanie

```bash
# Pierwsze logowanie — tworzy profil i cachuje tokeny
ksef-cli login --profile moja-firma --nip 1234567890 --token <token_ksef> --save-token

# Ponowne logowanie używa zapisanego tokenu KSeF
ksef-cli login --profile moja-firma
```

Token KSeF można wygenerować w portalu podatnika lub pobrać z konta biura rachunkowego.

### Listowanie faktur

```bash
# Lista faktur (ostatnie 3 miesiące)
ksef-cli invoice list --profile moja-firma

# Wyjście w formacie JSON
ksef-cli invoice list --profile moja-firma --output json

# Wybór okresu
ksef-cli invoice list --profile moja-firma --from 2026-02-01 --to 2026-03-31

# Wybór typu faktur
ksef-cli invoice list --profile moja-firma --subject subject2

# Paginacja
ksef-cli invoice list --profile moja-firma --pageoffset 2 --pagesize 50
```

Przykładowe wyjście tekstowe:
```
[torgiren@smartraptor ksef-cli (master)]$ ./ksef-cli invoice list --profile qwe 
+------------+--------------------------------------+------------+------------+-------------------------------------------------------------------------------------+
| DATA       | FAKTURA                              | BRUTTO     | NETTO      | KONTRAHENT                                                                          |
+------------+--------------------------------------+------------+------------+-------------------------------------------------------------------------------------+
| 2026-01-16 | FV26/01/2026                         | 40.00      | 32.52      | KM Hotel DEV 1                                                                      |
| 2026-01-16 | 1/HUZAR/2026                         | 20.00      | 20.00      | huzar                                                                               |
| 2026-01-16 | 5/HUZAR/2026                         | 20.00      | 20.00      | huzar                                                                               |
| 2026-01-16 | KOR/2/2026                           | 108.00     | 100.00     | Firma Przykladowa Sp. z o.o.                                                        |
| 2026-01-16 | KOR/3/2026                           | -1246.65   | -1000.00   | Firma Przykladowa Sp. z o.o.                                                        |
| 2026-01-18 | 17                                   | 615.00     | 500.00     | pokopoko                                                                            |
| 2026-01-18 | 18                                   | 1500.00    | 1500.00    | hokokokokh                                                                          |
| 2026-01-19 | 5/BA/2025                            | 861.00     | 700.00     | Jan Kowalski                                                                        |
| 2026-01-20 | E67C085C-7127-4501-B2B1-BC41F5B783D8 | 829.49     | 674.37     | STI DEV SPRZEDAWCA                                                                  |
| 2026-01-20 | FV/GWW/ENE/2024/09/073               | 1.23       | 1.00       | STI DEV SPRZEDAWCA                                                                  |
| 2026-01-20 | FV/GWW/ENE/2024/09/074               | 1.23       | 1.00       | STI DEV SPRZEDAWCA                                                                  |
| 2026-01-20 | FV/GWW/ENE/2024/09/075               | 1.23       | 1.00       | STI DEV SPRZEDAWCA                                                                  |
| 2026-01-21 | FV/000005/26                         | 136.53     | 111.00     | STUDIO AS TOMASZ SZULGA                                                             |
+------------+--------------------------------------+------------+------------+-------------------------------------------------------------------------------------+
```

### Flagi globalne

| Flaga | Domyślna wartość | Opis |
|-------|-----------------|------|
| `--profile` |  | Nazwa profilu do użycia |
| `--nip` |  | Nadpisuje NIP z profilu |
| `--output` | `text` | Format wyjścia: `text` lub `json` |
| `--test` | `false` | Używa środowiska testowego KSeF |
| `-a, --api` | `https://api.ksef.mf.gov.pl/v2` | Adres API KSeF |
| `--cache-dir` | `~/.cache/ksef-cli` | Katalog na cache tokenów |
| `--configFile` | `~/.config/ksef-cli/config.yaml` | Ścieżka do pliku konfiguracyjnego |
| `-v` |  | Poziom logowania: `-v` INFO, `-vv` DEBUG, `-vvv` TRACE, `-vvvv` SECRET |

### Środowisko testowe

```bash
ksef-cli login --test --profile test-firma --nip 1234567890 --token <token>
ksef-cli invoice list --test --profile test-firma
```

## Licencja

[GNU AGPL v3](LICENSE)
