package main

import (
    "fmt"
    "net/http"
    "os"
    "io"
    "strings"
    "time"
	"bufio"
	"encoding/json"
)

//
type Event struct {
	Name        string `json:"name"`
	Place       string `json:"place"`
	Description string `json:"description"`
	DtStart 	string `json:"dtstart"`
	DtEnd		string `json:"dtend"`
}

//

// Download and format .ics file function >>> file output readable by getEvents()
func getICS(url string, calendarsFolder string, filename string) (string, error) { // Function: args: (url in .ics, folder for saving in folder/ format, file name in filename.extension format)
    // Download specified .ics file
    response, err := http.Get(url)
    if err != nil {
        return "0", err
    }
    defer response.Body.Close()

    // Checking the response code status
    if response.StatusCode != http.StatusOK {
        return "0", fmt.Errorf("Erreur de téléchargement: Code de statut %d", response.StatusCode)
    }

	fullPath := calendarsFolder + filename

    // Local file creation
    file, err := os.Create(fullPath)
    if err != nil {
        return "0", err
    }
    defer file.Close()

    // Copy downloaded data to folder
    _, err = io.Copy(file, response.Body)
    if err != nil {
        return "0", err
    }

    // Current date in "YYYY-MM-DD" format
    currentDate := time.Now().Format("2006-01-02")

    // New file name using the specified format
    parts := strings.Split(filename, ".")
    if len(parts) < 2 {
        return "0", fmt.Errorf("Nom de fichier invalide")
    }
    newFilename := fmt.Sprintf("%s-%s.%s", strings.TrimSuffix(parts[0], "_"), currentDate, parts[1])

    // Renaming
    err = os.Rename(fullPath, calendarsFolder + newFilename)
    if err != nil {
        return "0", err
    }
	fmt.Println("newFilename name:", newFilename)
    return newFilename, nil
}

//

//
func getEvents(newFilename, calendarsFolder, startDate, endDate string) error {
    // Create of the complete path
    fullPath := calendarsFolder + newFilename

    // Open the file in read mode
    file, err := os.Open(fullPath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Create .json file without a .txt extension
    jsonFilename := strings.TrimSuffix(newFilename, ".txt") + ".json"

    // Create full path for .json file
    jsonFullPath := calendarsFolder + jsonFilename

    // Create a JSON file to write event data
    jsonFile, err := os.Create(jsonFullPath)
    if err != nil {
        return err
    }
    defer jsonFile.Close()

    // Crée un encodeur JSON pour écrire les données dans le fichier JSON
    encoder := json.NewEncoder(jsonFile)

    // Crée un slice pour stocker tous les événements
    var events []Event

    // Variables temporaires pour stocker les informations de l'événement en cours de traitement
    var currentEvent Event
    var currentKey string

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        switch {
        case strings.HasPrefix(line, "DTSTART:"):
            date := line[8:18] // Extrait la date de DTSTART au format YYYYMMDD
            if date >= startDate && date <= endDate {
                currentKey = date
            }
        case strings.HasPrefix(line, "DTEND:") && currentKey != "":
            date := line[6:16] // Extrait la date de DTEND au format YYYYMMDD
            currentKey += "-" + date

            // Ajoute l'événement au slice
            events = append(events, currentEvent)
        case currentKey != "":
            switch {
            case strings.HasPrefix(line, "SUMMARY:"):
                currentEvent.Name = line[8:]
            case strings.HasPrefix(line, "LOCATION:"):
                currentEvent.Place = line[9:]
            case strings.HasPrefix(line, "DESCRIPTION:"):
                currentEvent.Description = line[12:]
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return err
    }

    // Écrit tous les événements dans le fichier JSON
    if err := encoder.Encode(events); err != nil {
        return err
    }

    fmt.Printf("Les données ont été enregistrées dans le fichier JSON : %s\n", jsonFullPath)

    return nil
}

//

//
func CalendarGeneration() { // Calendar Generation function

}

//

//
func main() { // now for debogging 
    url := "https://edt.univ-nantes.fr/iut_nantes/g3173.ics" // url en .ics
    filename := "calendar.txt"                // nom du fichier (changeable par classe)
	calendarsFolder := "iCalendars/"	// chemin d'enregistrement des calendriers
	startDate := "20230901"
	endDate := "20230930"
	//picturesFolder := "calendars/"
	var newFilename string

    newFilename, err := getICS(url, calendarsFolder, filename)
    if err != nil {
        fmt.Println("Erreur:", err)
        return
    }

    fmt.Println("Téléchargement réussi et renommé:", newFilename)

	err = getEvents(newFilename, calendarsFolder, startDate, endDate)
	if err != nil {
		fmt.Println("Erreur:", err)
		return
	}
}