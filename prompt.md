
You are an expert Golang backend architect.
Generate a **production-ready Golang project** with the following requirements:

---

## 🛠 Functional Requirements (Business Logic)

1. **File Upload & Storage**

   * A Flask API (already exists) uploads `.xlsx` grade files into **S3 storage**.
   * Metadata about the uploaded file is inserted into a **SQL database** (`files` table) with status `UPLOADED`.
   * After upload, a message `{file_id, s3_path, school_id}` is pushed into a **Redis queue** for further processing.

2. **Ingestion Worker**

   * A Golang service consumes jobs from the ingestion queue.
   * Downloads the `.xlsx` file from S3.
   * Parses the Excel file and validates schema (must contain `student_id, subject, class, semester, grade`).
   * Inserts parsed data into a **staging table** (`grades_staging`) with `status=READY`.
   * If parsing fails → update file status to `PARSED_FAIL` with `error_message`.
   * If parsing succeeds → update file status to `PARSED_OK`.

3. **Sync Worker**

   * A separate Golang worker consumes jobs from a **sync queue**.
   * Jobs contain `{file_id, class, subject}`.
   * Worker queries the staging DB for matching records.
   * Performs batch insert into the **Education Department API** (external system).
   * Handles **auth tokens** (expire in 5 minutes, must auto-refresh before each batch).
   * Updates `grades_staging` records to `SYNCED` or `FAILED`.
   * Logs detailed errors for failed students.

4. **User Action**

   * On the frontend, after upload is parsed successfully, users can click a button **“Sync to Education Dept”**.
   * This triggers the API to:

     * Check whether the current time is within the allowed sync window.
     * If valid → enqueue a sync job `{file_id, class, subject}` into Redis.
     * If invalid → return error “Sync not allowed at this time”.

5. **Retry & Error Handling**

   * Network/API errors → retry automatically with exponential backoff (via Redis DLQ).
   * Business logic errors (invalid student\_id, grade out of range) → mark record as `FAILED` (no auto retry).
   * Provide an API `/grades/status/{file_id}` for checking sync results (success/fail counts, error messages).

---

## 🏗 Technical Requirements

* Language: **Golang 1.22+**.
* Project structure must follow **clean architecture / layered design**.
* Problem must solved:
   * Transaction when do upload and sync, ensures that it must be successful at all or rollback-able | notify user the status of the process.
   * Make sure that the processing time when syncing data is completed before the token/login expires.

* Must include:

  * `cmd/` → entrypoints for ingestion worker, sync worker.
  * `internal/` → core business logic.

    * `api/` → HTTP handlers (for status, trigger sync).
    * `db/` → database access layer (SQL queries).
    * `excel/` → Excel parser & validator.
    * `queue/` → Redis producer/consumer abstraction.
    * `sync/` → service to call Education Dept API.
    * `model/` → domain models & DTOs.
  * `config/` → YAML/JSON config (DB, Redis, S3, API endpoints, worker count).
* Use **worker pool pattern** for Redis consumers.
* Use **context with cancellation** for graceful shutdown.
* Apply **repository pattern** for DB access.
* Apply **strategy pattern** if multiple parsing strategies may exist (Excel now, CSV/PDF later).
* Logging: structured logging (zerolog or logrus).
* Error handling: clear error types, retry logic, DLQ support.
* Dependency management: Go modules.

---

## 📦 Output

* A **ready-to-run Golang project skeleton** with some runnable example code.
* Include:

  * `docker-compose.yml` (Postgres + Redis + MinIO as S3 mock).
  * `Dockerfile` for the Golang service.
  * Example config file `config.yaml`.
  * SQL migration scripts for `files` and `grades_staging`.
  * Sample Excel file (`grades.xlsx`).
* Ensure the project builds and runs with `docker-compose up --build`.

---

⚡ **Important**:
Follow Golang best practices, ensure **clean separation of concerns**, use **interfaces for external dependencies** (DB, S3, Redis, External API).
