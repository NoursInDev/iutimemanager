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
    "strconv"
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
            date := line[8:18] // Extrait DTSTART date au format YYYYMMDD
            fullDate := line[8:]
            if date >= startDate && date <= endDate {
                currentKey = date
                currentEvent.DtStart = fullDate // Ajoutez la date DTSTART à l'événement actuel
            }
        case strings.HasPrefix(line, "DTEND:") && currentKey != "":
            date := line[6:16] // Extrait DTEND date au format YYYYMMDD
            fullDate := line [6:]
            currentKey += "-" + date
            currentEvent.DtEnd = fullDate // Ajoutez la date DTEND à l'événement actuel

            // Ajoutez l'événement à la slice
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
    const width = 2560  //>>>config.json
    const height = 1440 //>>>config.json
    lineWidth := 5.0    //>>>config.json
    dc := gg.NewContext(width, height)
    dc.SetLineWidth(lineWidth)
    // Définir les couleurs à partir des chaînes hexadécimales
    mainHexColor := parseHexColor(mainColor)
    textHexColor := parseHexColor(textColor)
    markHexColor := parseHexColor(markColor)
    scdHexColor := parseHexColor(scdColor)

    // Remplir l'arrière-plan avec la couleur scdHexColor
    //dc.SetColor(markHexColor)
    //dc.Clear()
    fmt.Println(markHexColor)
    // Dessiner les barres horaires
    dc.SetColor(scdHexColor)
    x_offset := 0.1 * width
    for i := 8; i <= 20; i++ {
        y := float64(i-7) * height / 14 + lineWidth/2
        dc.DrawLine( 0, y, width, y)
        dc.Stroke()
    }
    for i:= 1; i <= 5; i++ {
        x := float64(i) * width / 5 * 0.9 + lineWidth/2 - (x_offset) - 50
        dc.DrawLine(x, 0, x, height)
        dc.Stroke()
    }

    // Dessiner les jours et les événements
    dc.SetColor(textHexColor)
    dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSerif-Bold.ttf", 12) // Spécifiez le chemin de votre police de caractères

    for _, event := range events {
        // Dessiner le rectangle de l'événement avec la couleur principale
        DtStart_data := event.DtStart
        DtEnd_data := event.DtEnd

        dtstart_date_data, _ := extractDate(DtStart_data)
        dtstart_hour_data, _ := extractHour(DtStart_data)
        dtend_hour_data, _ := extractHour(DtEnd_data)

        x1_placement_var, err := getDayOfWeek(dtstart_date_data)
        if err != nil {
            fmt.Println("Erreur:", err)
        }
        y1_placement_var, err := timeStringToHours(dtstart_hour_data)
        if err != nil {
            fmt.Println("Erreur:", err)
        }
        y2_placement_var, err := timeStringToHours(dtend_hour_data)
        if err != nil {
            fmt.Println("Erreur:", err)
        }

        x1_placement := ((x1_placement_var - 1) * width / 5) + (x_offset) - 50*x1_placement_var
        x2_placement := float64(width / 5) - 150
        y1_placement := (y1_placement_var - 7) * height / 14
        y2_placement := 1.33 * height / 14
        fmt.Println(y1_placement, y2_placement_var)


        dc.SetColor(mainHexColor)
        dc.DrawRectangle(x1_placement, y1_placement, x2_placement, y2_placement) // Spécifiez les coordonnées et les dimensions de votre rectangle
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
func loadEvents(filename string) ([]Event, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var events []Event
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&events)
    if err != nil {
        return nil, err
    }

    return events, nil
}

func extractDate(input string) (string, error) {
    // Définir le format de la chaîne d'entrée
    layout := "20060102T150405Z"

    // Analyser la chaîne d'entrée en tant que temps
    t, err := time.Parse(layout, input)
    if err != nil {
        return "", err
    }

    // Extraire la date au format YYYYMMDD
    dateStr := t.Format("20060102")

    return dateStr, nil
}

func extractHour(input string) (string, error) {
    // Définir le format de la chaîne d'entrée
    layout := "20060102T150405Z"

    // Analyser la chaîne d'entrée en tant que temps
    t, err := time.Parse(layout, input)
    if err != nil {
        return "", err
    }

    // Ajouter +2 heures au temps
    t = t.Add(2 * time.Hour)

    // Extraire l'heure au format HHMMSS
    hourStr := t.Format("150405")

    return hourStr, nil
}

func getDayOfWeek(inputDate string) (float64, error) {
    // Convertir la chaîne d'entrée au format "YYYYMMDD" en une valeur de type time.Time
    date, err := time.Parse("20060102", inputDate)
    if err != nil {
        return 0, err
    }

    // Obtenir le jour de la semaine (0 = dimanche, 1 = lundi, ..., 6 = samedi)
    dayOfWeek := float64(date.Weekday())

    // Remapper le dimanche de 0 à 7
    if dayOfWeek == 0 {
        dayOfWeek = 7
    }

    return dayOfWeek, nil
}

func timeStringToHours(input string) (float64, error) {
    // Assurez-vous que la chaîne d'entrée a une longueur valide (HHMMSS)
    if len(input) != 6 {
        return 0, fmt.Errorf("Problème de la chaine d'entrée: pas sous format (HHMMSS)")
    }

    // Extraire les heures, les minutes et les secondes de la chaîne d'entrée
    hoursStr := input[0:2]
    minutesStr := input[2:4]
    secondsStr := input[4:6]

    // Convertir les parties de la chaîne en entiers
    hours, err := strconv.Atoi(hoursStr)
    if err != nil {
        return 0, err
    }

    minutes, err := strconv.Atoi(minutesStr)
    if err != nil {
        return 0, err
    }

    seconds, err := strconv.Atoi(secondsStr)
    if err != nil {
        return 0, err
    }

    // Calculer le nombre total de minutes
    totalHours := float64(hours + minutes/60) + float64(seconds)/3600.0

    return totalHours, nil
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