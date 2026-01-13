package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jung-kurt/gofpdf"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Adatbázis kapcsolat globális kezelése
func getDB() (*sql.DB, error) {
	user := getEnv("DB_USER", "glpi")
	pass := getEnv("DB_PASSWORD", "glpi")
	name := getEnv("DB_NAME", "glpi")
	host := getEnv("DB_HOST", "192.168.1.253")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4", user, pass, host, name)
	return sql.Open("mysql", dsn)
}

func main() {
	// Kezdőoldal: Egyszerű link lista
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `
			<h1>Riportok</h1>
			<ul>
				<li><a href="/riport/gepek">Számítógépek és licenszek (PDF)</a></li>
				</ul>
		`)
	})

	// PDF riport útvonala
	http.HandleFunc("/riport/gepek", handleAssetReport)

	port := "5001"
	fmt.Printf("Szerver elindult a http://localhost:%s címen...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleAssetReport(w http.ResponseWriter, r *http.Request) {
	db, err := getDB()
	if err != nil {
		http.Error(w, "Adatbázis hiba", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Lekérdezés
	query := `
		SELECT
			COALESCE(c.name, '') as computer_name,
			COALESCE(l.completename, 'Nincs megadva') as location,
			COALESCE(os.name, 'Ismeretlen') as os_name,
			COALESCE(osv.name, '-') as os_version,
			COALESCE(ios.license_number, '-') as license
		FROM glpi_computers c
		LEFT JOIN glpi_locations l ON c.locations_id = l.id
		LEFT JOIN glpi_items_operatingsystems ios ON c.id = ios.items_id AND ios.itemtype = 'Computer'
		LEFT JOIN glpi_operatingsystems os ON ios.operatingsystems_id = os.id
		LEFT JOIN glpi_operatingsystemversions osv ON ios.operatingsystemversions_id = osv.id
		WHERE c.is_deleted = 0
		ORDER BY c.name ASC`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Lekérdezési hiba", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// PDF generálása
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddUTF8Font("IR-Reg", "", "Inter-Regular.ttf")
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()

	// Fejléc
	pdf.SetFont("IR-Reg", "", 12)
	pdf.CellFormat(0, 10, "Számítógépek és licenszek", "", 1, "C", false, 0, "")
	pdf.SetFont("IR-Reg", "", 8)
	pdf.CellFormat(0, 0, fmt.Sprintf("Készült: %s", time.Now().Format("2006.01.02. 15:04")), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Táblázat fejléc
	cols := []float64{12, 43, 55, 50, 40, 75}
	headers := []string{"#", "Gép neve", "Helyszín", "OS", "Verzió", "Licensz"}
	pdf.SetFillColor(241, 243, 245)
	pdf.SetFont("IR-Reg", "", 8)
	for i, h := range headers {
		pdf.CellFormat(cols[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Adatok
	pdf.SetFont("IR-Reg", "", 7.5)
	count := 1
	alternate := false
	for rows.Next() {
		var name, loc, osName, osVer, license string
		if err := rows.Scan(&name, &loc, &osName, &osVer, &license); err != nil {
			continue
		}

		if alternate {
			pdf.SetFillColor(248, 249, 250)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		pdf.CellFormat(cols[0], 7, fmt.Sprintf("%d.", count), "1", 0, "C", true, 0, "")
		pdf.CellFormat(cols[1], 7, name, "1", 0, "L", true, 0, "")
		pdf.CellFormat(cols[2], 7, loc, "1", 0, "L", true, 0, "")
		pdf.CellFormat(cols[3], 7, osName, "1", 0, "L", true, 0, "")
		pdf.CellFormat(cols[4], 7, osVer, "1", 0, "L", true, 0, "")
		pdf.CellFormat(cols[5], 7, license, "1", 1, "L", true, 0, "")

		alternate = !alternate
		count++
	}

	// Válasz beállítása letöltésként
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=gep_riport.pdf")

	err = pdf.Output(w) // Közvetlenül a ResponseWriter-be írjuk a PDF-et
	if err != nil {
		log.Printf("Hiba a PDF küldésekor: %v", err)
	}
}
