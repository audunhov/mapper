# Web-Based Mapping UI (File Upload & Header Mapping)

In many web applications, users upload spreadsheets (CSV or XLSX) with their own custom header names and map them to your system's database/struct fields.

A standard, stateless pattern to achieve this is:
1. **Upload & Parse Headers**: The user selects a file. The frontend uploads it to the server (e.g. `POST /parse-headers`), which parses only the headers and returns them as a JSON list.
2. **Interactive Mapping UI**: The frontend displays target fields next to dropdown selectors populated with the file's headers, automatically selecting the closest matches.
3. **Submit Mapping & Import**: When the user clicks "Import", the frontend uploads the same file again, along with the user's header-to-field mapping config as a JSON string. The server converts the rows using `mapper.Convert` and the custom mapping.

Below is a complete, self-contained Go web server showing this full flow (HTML/CSS/JS frontend included):

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/audunhov/mapper"
)

// Contact is our target database schema
type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Age   int    `json:"age"`
}

// targetFields lists the exact struct field names we want mapped in the UI
var targetFields = []string{"Name", "Email", "Phone", "Age"}

func main() {
	// Serve the frontend page
	http.HandleFunc("/", handleIndex)

	// Endpoint to parse headers from the uploaded file
	http.HandleFunc("/parse-headers", handleParseHeaders)

	// Endpoint to convert the file using the custom mapping
	http.HandleFunc("/import", handleImport)

	fmt.Println("Server starting at http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleIndex serves the HTML page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

// handleParseHeaders reads the uploaded file's headers and returns them as JSON
func handleParseHeaders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var importer mapper.Importer
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == ".xlsx" {
		importer = mapper.NewXLSXImporter(file, "")
	} else {
		importer = mapper.NewCSVImporter(file)
	}

	headers, err := importer.ReadHeaders()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read headers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"headers": headers,
	})
}

// handleImport converts the uploaded file into Go structs using the user's custom ImportMap
func handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse custom mapping from request FormValue
	mappingStr := r.FormValue("mapping")
	var mapping mapper.ImportMap
	if err := json.Unmarshal([]byte(mappingStr), &mapping); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse mapping: %v", err), http.StatusBadRequest)
		return
	}

	var importer mapper.Importer
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == ".xlsx" {
		importer = mapper.NewXLSXImporter(file, "")
	} else {
		importer = mapper.NewCSVImporter(file)
	}

	imported, err := importer.Import()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to import data: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply target mapping
	contacts, err := mapper.Convert[Contact](imported, mapping)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"count":   len(contacts),
		"data":    contacts,
	})
}

const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CSV/XLSX Custom Mapper</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
            background-color: #f9fafb;
            color: #1f2937;
            max-width: 750px;
            margin: 40px auto;
            padding: 0 20px;
            line-height: 1.5;
        }
        .card {
            background: #ffffff;
            border: 1px solid #e5e7eb;
            border-radius: 8px;
            padding: 24px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.05);
            margin-bottom: 24px;
        }
        h1, h2 {
            margin-top: 0;
            color: #111827;
        }
        .btn {
            background-color: #2563eb;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            font-weight: 500;
            cursor: pointer;
        }
        .btn:hover {
            background-color: #1d4ed8;
        }
        .btn-secondary {
            background-color: #4b5563;
        }
        .btn-secondary:hover {
            background-color: #374151;
        }
        .mapping-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #f3f4f6;
        }
        .mapping-row select {
            width: 250px;
            padding: 8px;
            border-radius: 6px;
            border: 1px solid #d1d5db;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 16px;
        }
        th, td {
            text-align: left;
            padding: 12px;
            border-bottom: 1px solid #e5e7eb;
        }
        th {
            background-color: #f3f4f6;
            font-weight: 600;
        }
        .hidden {
            display: none;
        }
    </style>
</head>
<body>

    <h1>CSV/XLSX Data Importer</h1>
    <p>Upload a file, dynamically map fields, and convert using Go and the ` + "`" + `mapper` + "`" + ` library.</p>

    <!-- Step 1: Choose File -->
    <div id="step-upload" class="card">
        <h2>Step 1: Choose File</h2>
        <div style="margin-bottom: 16px;">
            <input type="file" id="file-input" accept=".csv,.xlsx" />
        </div>
        <button class="btn" onclick="uploadAndParseHeaders()">Next: Map Columns</button>
    </div>

    <!-- Step 2: Map Headers -->
    <div id="step-mapping" class="card hidden">
        <h2>Step 2: Map Columns</h2>
        <p>Align the column headers from your file with our system's fields:</p>
        <div id="mapping-fields"></div>
        <div style="margin-top: 24px; display: flex; gap: 12px;">
            <button class="btn btn-secondary" onclick="backToUpload()">Back</button>
            <button class="btn" onclick="importData()">Import & Convert</button>
        </div>
    </div>

    <!-- Step 3: Results -->
    <div id="step-results" class="card hidden">
        <h2>Step 3: Import Summary</h2>
        <p id="result-summary" style="font-weight: 600; color: #16a34a;"></p>
        <div style="overflow-x: auto;">
            <table id="result-table">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Email</th>
                        <th>Phone</th>
                        <th>Age</th>
                    </tr>
                </thead>
                <tbody id="result-body"></tbody>
            </table>
        </div>
        <div style="margin-top: 24px;">
            <button class="btn" onclick="resetForm()">Import Another File</button>
        </div>
    </div>

    <script>
        const targetFields = ["Name", "Email", "Phone", "Age"];
        let fileToUpload = null;

        async function uploadAndParseHeaders() {
            const input = document.getElementById("file-input");
            if (!input.files || input.files.length === 0) {
                alert("Please select a CSV or XLSX file.");
                return;
            }
            fileToUpload = input.files[0];

            const formData = new FormData();
            formData.append("file", fileToUpload);

            try {
                const response = await fetch("/parse-headers", {
                    method: "POST",
                    body: formData
                });
                if (!response.ok) throw new Error(await response.text());
                
                const data = await response.json();
                renderMappingFields(data.headers);
            } catch (err) {
                alert("Failed to parse headers: " + err.message);
            }
        }

        function renderMappingFields(headers) {
            const container = document.getElementById("mapping-fields");
            container.innerHTML = "";

            targetFields.forEach(field => {
                const row = document.createElement("div");
                row.className = "mapping-row";

                const label = document.createElement("span");
                label.style.fontWeight = "600";
                label.textContent = field;

                const select = document.createElement("select");
                select.id = "target-" + field;

                const defaultOpt = document.createElement("option");
                defaultOpt.value = "";
                defaultOpt.textContent = "-- Skip Column --";
                select.appendChild(defaultOpt);

                headers.forEach(h => {
                    const opt = document.createElement("option");
                    opt.value = h;
                    opt.textContent = h;
                    select.appendChild(opt);
                });

                // Auto-match if header closely matches field name (case-insensitive fuzzy)
                const bestMatch = headers.find(h => {
                    const hClean = h.toLowerCase().replace(/[^a-z0-9]/g, "");
                    const fClean = field.toLowerCase().replace(/[^a-z0-9]/g, "");
                    return hClean.includes(fClean) || fClean.includes(hClean);
                });
                if (bestMatch) select.value = bestMatch;

                row.appendChild(label);
                row.appendChild(select);
                container.appendChild(row);
            });

            document.getElementById("step-upload").classList.add("hidden");
            document.getElementById("step-mapping").classList.remove("hidden");
        }

        async function importData() {
            const mapping = {};
            targetFields.forEach(field => {
                const val = document.getElementById("target-" + field).value;
                if (val) {
                    // map file header (val) -> struct field name (field)
                    mapping[val] = field;
                }
            });

            const formData = new FormData();
            formData.append("file", fileToUpload);
            formData.append("mapping", JSON.stringify(mapping));

            try {
                const response = await fetch("/import", {
                    method: "POST",
                    body: formData
                });
                if (!response.ok) throw new Error(await response.text());

                const res = await response.json();
                showResults(res);
            } catch (err) {
                alert("Failed to import data: " + err.message);
            }
        }

        function showResults(res) {
            document.getElementById("step-mapping").classList.add("hidden");
            document.getElementById("step-results").classList.remove("hidden");
            document.getElementById("result-summary").textContent = "Successfully converted " + res.count + " records.";

            const tbody = document.getElementById("result-body");
            tbody.innerHTML = "";
            res.data.forEach(item => {
                const tr = document.createElement("tr");
                tr.innerHTML = `
                    <td>${item.name || ""}</td>
                    <td>${item.email || ""}</td>
                    <td>${item.phone || ""}</td>
                    <td>${item.age || 0}</td>
                `;
                tbody.appendChild(tr);
            });
        }

        function backToUpload() {
            document.getElementById("step-mapping").classList.add("hidden");
            document.getElementById("step-upload").classList.remove("hidden");
        }

        function resetForm() {
            fileToUpload = null;
            document.getElementById("file-input").value = "";
            document.getElementById("step-results").classList.add("hidden");
            document.getElementById("step-upload").classList.remove("hidden");
        }
    </script>
</body>
</html>
`
```

---

## Production Tips for Web-Based Imports

When building production-ready file importers for web applications, handling file uploads statelessly (as shown in the basic example) can lead to performance issues, high memory consumption, or poor user experience. Here are critical tips and best practices:

### 1. File Caching and Session-Based Temporary Storage
In the simple example, the user uploads the entire file twice:
* First, to `POST /parse-headers` (to read header names).
* Second, to `POST /import` (to process headers alongside the mapping definition).

For large files (e.g., > 5MB), double uploading wastes bandwidth and processing time. 
* **Upload Once with Session Token**: On the first upload, save the file to a temporary directory on the server (using Go's `os.CreateTemp` or a dedicated folder structure) or an object store (like AWS S3 / MinIO). Generate a unique, randomly-generated `Session ID` (UUID) or use the user's session token.
* **Return Headers & ID**: Return the extracted headers list *and* the `Session ID` to the frontend.
* **Submit Mapping via Session ID**: When importing, the frontend only sends the `Session ID` and the JSON mapping config. The server loads the file from the local disk/object store using the `Session ID`, parses/processes it, and then deletes the temporary file.
* **Clean-up Worker**: Implement a background cron job (or Go goroutine with a ticker) that deletes files older than 1–2 hours in the temporary directory to avoid running out of disk space from abandoned uploads.

### 2. Memory Management: Streaming vs. In-Memory Uploads
By default, standard libraries read uploads into memory if they fit under the specified multipart form memory threshold.
* **Streaming Files**: For large files, stream the request body directly to a temporary file on disk rather than caching the whole file structure in RAM. Use `r.MultipartReader()` instead of `r.ParseMultipartForm()` for strict stream processing, or adjust the memory limit in `r.ParseMultipartForm(MaxMemory)` to a sensible minimum (like 1–2 MB).
* **Garbage Collection Optimization**: Reading files into huge byte slices causes high heap allocation and triggers frequent GC sweeps. Streaming directly to disk ensures minimal RAM overhead.

### 3. Background Processing & Queues for Large Files
Importing files with tens of thousands of rows takes time. If a user imports a 50MB spreadsheet, a synchronous HTTP request will time out (most reverse proxies like Nginx/Cloudflare terminate requests after 30 to 60 seconds).
* **Asynchronous Jobs**: When receiving the `/import` request, validate the mapping, write a status database record (e.g., `Status: Processing`), enqueue an import job in a queue (e.g., Redis/RabbitMQ or a pool of Go worker goroutines), and immediately return a `202 Accepted` response with the Job ID.
* **Status Updates**: The frontend can poll `/import-status/:job_id` or establish a WebSocket connection to receive progress percentage updates (e.g., "Row 15000 of 50000 processed").
* **Chunking Database Inserts**: When converting the imported rows into structs, do not perform single SQL insert queries for each row. Instead, write rows to the database in batches (e.g., 500–1000 records per transaction) to drastically speed up database write performance.

### 4. Excel (XLSX) Memory Constraints
Excel files are zipped XML documents and are notoriously memory-intensive to parse compared to flat CSVs.
* **xlsx Library Overhead**: Most Go Excel libraries load the entire spreadsheet into memory. When processing very large `.xlsx` files, consider using libraries that support streaming reading (like `excelize`'s `Rows` iterator).
* **Enforcing Limits**: Set a lower size limit for Excel uploads (e.g., 20MB) than CSV uploads (e.g., 100MB).
