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
    "image/color"
    "github.com/fogleman/gg"
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

    // Check the response code status
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

    // Download copy data to folder
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

    // Rename file
    err = os.Rename(fullPath, calendarsFolder + newFilename)
    if err != nil {
        return "0", err
    }
	fmt.Println("newFilename name:", newFilename)
    return newFilename, nil
}

//

//
func getEvents(newFilename, calendarsFolder, startDate, endDate string) (string, error) {
    // Create of the complete path
    fullPath := calendarsFolder + newFilename

    // Open the file in read mode
    file, err := os.Open(fullPath)
    if err != nil {
        return "0", err
    }
    defer file.Close()

    // Create .json file without a .txt extension
    jsonFilename := strings.TrimSuffix(newFilename, ".txt") + ".json"

    // Create full path for .json file
    jsonFullPath := calendarsFolder + jsonFilename

    // Create a JSON file to write event data
    jsonFile, err := os.Create(jsonFullPath)
    if err != nil {
        return "0", err
    }
    defer jsonFile.Close()

    // Create a JSON encoder to write data to the JSON file
    encoder := json.NewEncoder(jsonFile)

    // Create a slice to store all events
    var events []Event

    // Temporary variables to store event information currently being processed
    var currentEvent Event
    var currentKey string

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        switch {
        case strings.HasPrefix(line, "DTSTART:"):
            date := line[8:18] // Extract DTSTART date in YYYYMMDD format
            if date >= startDate && date <= endDate {
                currentKey = date
            }
        case strings.HasPrefix(line, "DTEND:") && currentKey != "":
            date := line[6:16] // Extract DTEND date in YYYYMMDD format
            currentKey += "-" + date

            // Add event to slice
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
        return "0", err
    }

    // Write all events to JSON file
    if err := encoder.Encode(events); err != nil {
        return "0", err
    }

    fmt.Printf("Les données ont été enregistrées dans le fichier JSON : %s\n", jsonFullPath)

    return jsonFullPath, nil
}

//

//
func CalendarGeneration(picturesFolder, newJSONname, mainColor, textColor, markColor, scdColor, newFilename string) error {
    // Charger les données du fichier JSON (vous devrez implémenter cette partie)
    events, err := loadEvents(newJSONname)
    if err != nil {
        return err
    }

    // Créer une image avec une taille spécifique (vous pouvez ajuster la taille selon vos besoins)
    const width = 800
    const height = 600
    dc := gg.NewContext(width, height)

    // Définir les couleurs à partir des chaînes hexadécimales
    mainHexColor := parseHexColor(mainColor)
    textHexColor := parseHexColor(textColor)
    markHexColor := parseHexColor(markColor)
    scdHexColor := parseHexColor(scdColor)

    // Remplir l'arrière-plan avec la couleur scdHexColor
    dc.SetColor(scdHexColor)
    dc.Clear()

    // Dessiner les barres horaires
    dc.SetColor(markHexColor)
    for i := 7; i <= 21; i++ {
        x := float64(i-7) * width / 14
        dc.DrawLine(x, 0, x, height)
        dc.Stroke()
    }

    // Dessiner les jours et les événements
    dc.SetColor(textHexColor)
    dc.LoadFontFace("/path/to/your/font.ttf", 12) // Spécifiez le chemin de votre police de caractères

    for _, event := range events {
        // Dessiner le rectangle de l'événement avec la couleur principale
        dc.SetColor(mainHexColor)
        dc.DrawRectangle(100, 100, 200, 100) // Spécifiez les coordonnées et les dimensions de votre rectangle
        dc.Fill()

        // Écrire le nom, le lieu et la description de l'événement avec la couleur du texte
        dc.DrawStringWrapped(event.Name+"\n"+event.Place+"\n"+event.Description, 100, 100, 0, 0, 200, 12, gg.AlignLeft)
        dc.Stroke()
    }

    // Enregistrez l'image dans le dossier spécifié avec le nom spécifié (en supprimant .txt)
    imageFilename := picturesFolder + strings.TrimSuffix(newFilename, ".txt") + ".png"
    if err := dc.SavePNG(imageFilename); err != nil {
        return err
    }

    fmt.Printf("Calendrier généré et enregistré sous: %s\n", imageFilename)

    return nil
}

// Fonction pour charger les événements depuis le fichier JSON (à implémenter)
func loadEvents(newJSONname string) ([]Event, error) {
    // Implémentez la logique pour charger les événements à partir du fichier JSON
    // et retournez-les sous forme de slice d'événements
    return nil, nil
}

// Fonction pour analyser une couleur hexadécimale au format #RRGGBB
func parseHexColor(hexColor string) color.RGBA {
    hex := strings.TrimPrefix(hexColor, "#")
    r := hexToDec(hex[0:2])
    g := hexToDec(hex[2:4])
    b := hexToDec(hex[4:6])
    return color.RGBA{r, g, b, 255}
}

// Fonction pour convertir une chaîne hexadécimale en entier décimal
func hexToDec(hex string) uint8 {
    var result uint64
    for i := 0; i < len(hex); i++ {
        char := hex[i]
        digit := uint8(char - '0')
        if char >= 'a' && char <= 'f' {
            digit = uint8(char-'a') + 10
        } else if char >= 'A' && char <= 'F' {
            digit = uint8(char-'A') + 10
        }
        result = result*16 + uint64(digit)
    }
    return uint8(result)
}


//

//
func main() { // now for debogging 
    url := "https://edt.univ-nantes.fr/iut_nantes/g3173.ics"        // url in .ics format (>config.json)
    filename := "calendar.txt"                                      // file name configuration (>config.json)
	calendarsFolder := "iCalendars/"	                            // data file storage folder (>config.json)
	startDate := "20230901"                                         // week start date
	endDate := "20230930"                                           // week end date
	picturesFolder := "calendars/"                                  // calendar picture storage folder
    mainColor := "#EA94E2"                                          // main color       (HEXA)
    scdColor := "#F6928F"                                           // secondary color  (HEXA)
    textColor := "#FFFFFF"                                          // text color       (HEXA)
    markColor := "#FFE6FF"                                          // mark color       (HEXA)
	var newFilename string
    var newJSONname string

    newFilename, err := getICS(url, calendarsFolder, filename)
    if err != nil {
        fmt.Println("Erreur:", err)
        return
    }

    fmt.Println("Téléchargement réussi et renommé:", newFilename)

	newJSONname, err = getEvents(newFilename, calendarsFolder, startDate, endDate)
	if err != nil {
		fmt.Println("Erreur:", err)
		return
    	}
    err = CalendarGeneration(picturesFolder, newJSONname, mainColor, textColor, markColor, scdColor, newFilename)
    if err != nil {
        fmt.Println("Erreur:", err)
    }
}