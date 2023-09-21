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
    "io/ioutil"
)

//
type Event struct {
	Name        string `json:"name"`
	Place       string `json:"place"`
	Description string `json:"description"`
	DtStart 	string `json:"dtstart"`
	DtEnd		string `json:"dtend"`
    UID         string `json:"uid`
    Type_start  string `json:"type_start"`
    Type_end    string `json:"type_end"`
    Categories  string `json:"categories"`
    DtStamp     string `json:"dtstamp"`
}

type Config struct {
    Planning map[string]string `json:"planning"`
    Settings map[string]string `json:"generator_config"`
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
        return "0", fmt.Errorf("Download Error: Status Indicator: %d", response.StatusCode)
    }

	fullPath := calendarsFolder + filename + "_" + os.Args[1]

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
        return "0", fmt.Errorf("Invalid file name")
    }
    newFilename := fmt.Sprintf("%s-%s_%s.txt", filename, currentDate, os.Args[1])

    // Rename file
    err = os.Rename(fullPath, calendarsFolder + newFilename)
    if err != nil {
        return "0", err
    }
	fmt.Println("newFilename name:", newFilename)

    fmt.Println("Successfully downloaded and renamed:", newFilename)

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
    var currentEvent Event
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        switch {
            case strings.HasPrefix(line, "BEGIN:VEVENT"):
                currentEvent := Event{}
                currentEvent.Type_start = line [6:]
            case strings.HasPrefix(line, "DTSTART:"):
                currentEvent.DtStart = line [8:]
            case strings.HasPrefix(line, "DTEND:"):
                currentEvent.DtEnd = line [6:]
            case strings.HasPrefix(line, "UID:"):
                currentEvent.UID = line [4:]
            case strings.HasPrefix(line, "SUMMARY:"):
                currentEvent.Name = line [8:]
            case strings.HasPrefix(line, "LOCATION:"):
                currentEvent.Place = line [9:]
            case strings.HasPrefix(line, "DESCRIPTION:"):
                currentEvent.Description = line [12:]
            case strings.HasPrefix(line, "CATEGORIES"):
                currentEvent.Categories = line [11:]
            case strings.HasPrefix(line, "END:VEVENT"):
                currentEvent.Type_end = line [4:]
                events = append(events, currentEvent)
        }
    }

    if err := scanner.Err(); err != nil {
        return "0", err
    }

    if err := encoder.Encode(events); err != nil {
        return "0", err
    }

    
    fmt.Printf("Successfully saved in the JSON file: %s\n", jsonFullPath)

    return jsonFullPath, nil
}

//

//
func filterEventsByDateRange(jsonFilename, startDate, endDate string) error {
    // Load events from JSON file
    events, err := loadEvents(jsonFilename)
    if err != nil {
        return err
    }

    // Create a new list to store filtered events
    var filteredEvents []Event

    // Browse events and filter those within the specified date range
    for _, event := range events {
        eventStartDate, err := extractDate(event.DtStart)
        if err != nil {
            return err
        }

        // Check if the event is between startDate and endDate
        if eventStartDate >= startDate && eventStartDate <= endDate {
            filteredEvents = append(filteredEvents, event)
        }
    }

    // Write the filtered list to the JSON file
    jsonFile, err := os.Create(jsonFilename)
    if err != nil {
        return err
    }
    defer jsonFile.Close()

    encoder := json.NewEncoder(jsonFile)
    if err := encoder.Encode(filteredEvents); err != nil {
        return err
    }

    fmt.Printf("Filtered events successfully saved in JSON file: %s\n", jsonFilename)
    return nil
}

//

//
func CalendarGeneration(picturesFolder, newJSONname, mainColor, textColor, scdColor, newFilename string) error {
    // Load data from JSON file
    events, err := loadEvents(newJSONname)
    if err != nil {
        return err
    }

    // Create an image with a specific size
    const width = 2560  
    const height = 1440 
    lineWidth := 5.0    
    dc := gg.NewContext(width, height)
    dc.SetLineWidth(lineWidth)
    // Defines colors from hexadecimal strings
    mainHexColor := parseHexColor(mainColor)
    textHexColor := parseHexColor(textColor)
    scdHexColor := parseHexColor(scdColor)

    // Draw time bars
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

    // Draw days and events
    dc.SetColor(mainHexColor)
    dc.LoadFontFace("fonts/kanit/Kanit-Medium.ttf", 12)

    var arused_data [][]float64

    for _, event := range events {
        // Draw event rectangle with main color
        DtStart_data := event.DtStart

        dtstart_date_data, _ := extractDate(DtStart_data)
        dtstart_hour_data, _ := extractHour(DtStart_data)

        fmt.Println(dtstart_date_data, "***", dtstart_hour_data, "***", DtStart_data)

        x1_placement_var, err := getDayOfWeek(dtstart_date_data)
        if err != nil {
            fmt.Println("Error: Failed to retrieve the day: ", err)
        }
        y1_placement_var, err := timeStringToHours(dtstart_hour_data)
        if err != nil {
            fmt.Println("Erreur:", err)
        }

        x1_placement := ((x1_placement_var - 1) * width / 5) + (x_offset) - 50*x1_placement_var
        x2_placement := float64(width / 5) - 150
        y1_placement := (y1_placement_var - 7) * height / 14
        y2_placement := 1.33 * height / 14

        appendlist := []float64{x1_placement,x2_placement,y1_placement,y2_placement}
        
        if !contains(arused_data, appendlist) {
            arused_data = append(arused_data, appendlist)
            
            fmt.Println(x1_placement, " - ", x2_placement, " | ", y1_placement, " - ",y2_placement, " | ", y1_placement_var)

            dc.SetColor(mainHexColor)
            dc.DrawRectangle(x1_placement, y1_placement, x2_placement, y2_placement) // Specifying rectangle coordinates and dimensions
            dc.Fill()

            dc.SetColor(textHexColor)
            
            textX := x1_placement + 10
            textY := y1_placement + 10

            // Write the name, location and description of the event in text color
            dc.DrawStringWrapped(event.Name, textX, textY, 0, 0, 340, 9, gg.AlignLeft)
            dc.DrawStringWrapped(event.Place, textX + 10, textY + 10, 0, 0, 340, 9, gg.AlignLeft)
            dc.DrawStringWrapped(event.Description, textX + 20, textY + 20, 0, 0, 340, 9, gg.AlignLeft)
            dc.Stroke()
        }

    }

    // Save the image in the specified folder with the specified name (deleting .txt)
    imageFilename := picturesFolder + strings.TrimSuffix(newFilename, ".txt") + ".png"
    if err := dc.SavePNG(imageFilename); err != nil {
        return err
    }
    fmt.Printf("Calendar successfully generated and saved as: %s\n", imageFilename)

    return nil
}








// Function to load events from JSON file
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

// Function to extract Date from full Date&Time string
func extractDate(input string) (string, error) {
    // Define input string format
    layout := "20060102T150405Z"

    t, err := time.Parse(layout, input)
    if err != nil {
        return "", err
    }

    dateStr := t.Format("20060102")

    return dateStr, nil
}

// Function to extract Time (in Hours) from full Date&Time string
func extractHour(input string) (string, error) {
    // Define input string format
    layout := "20060102T150405Z"

    t, err := time.Parse(layout, input)
    if err != nil {
        return "", err
    }

    t = t.Add(2 * time.Hour)

    hourStr := t.Format("150405")

    return hourStr, nil
}

// Function to get the day in the week according to an input date
func getDayOfWeek(inputDate string) (float64, error) {
    // Convert the input string in "YYYYMMDD" format into a time.Time value
    date, err := time.Parse("20060102", inputDate)
    if err != nil {
        return 0, err
    }

    dayOfWeek := float64(date.Weekday())

    // Reset Sunday from 0 to 7 (EU format)
    if dayOfWeek == 0 {
        dayOfWeek = 7
    }

    return dayOfWeek, nil
}

func timeStringToHours(input string) (float64, error) {
    // check validity of input
    if len(input) != 6 {
        return 0, fmt.Errorf("Input chain problem: not in format (HHMMSS)")
    }

    hoursStr := input[0:2]
    minutesStr := input[2:4]
    secondsStr := input[4:6]

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

    totalHours := float64(hours) + float64(minutes)/60 + float64(seconds)/3600.0
    fmt.Println(totalHours)
    return totalHours, nil
}


// Function to analyze a hexadecimal color in #RRGGBB format
func parseHexColor(hexColor string) color.RGBA {
    hex := strings.TrimPrefix(hexColor, "#")
    r := hexToDec(hex[0:2])
    g := hexToDec(hex[2:4])
    b := hexToDec(hex[4:6])
    return color.RGBA{r, g, b, 255}
}

// Function to convert a hexadecimal string into a decimal integer
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

func areValuesEqual(values1, values2 []float64) bool {
    if len(values1) != len(values2) {
        return false
    }
    for i, value := range values1 {
        if value != values2[i] {
            return false
        }
    }
    return true
}

func contains(slice [][]float64, element []float64) bool {
    for _, item := range slice {
        if areValuesEqual(item, element) {
            return true
        }
    }
    return false
}



func areFirstEightDigitsEqual(dtstart, dtend string) bool {
	startTime, err := time.Parse("20060102T150405Z", dtstart)
	if err != nil {
		return false
	}
	endTime, err := time.Parse("20060102T150405Z", dtend)
	if err != nil {
		return false
	}

	return startTime.Format("20060102") == endTime.Format("20060102")
}



//

//
func main() {
    
    configFilename := "config.json"
    var config Config
    configFile, err := ioutil.ReadFile(configFilename)
    if err != nil {
        fmt.Println("Error reading config.json file:", err)
        return
    }

    // Decode the JSON content in the configuration structure
    if err := json.Unmarshal(configFile, &config); err != nil {
        fmt.Println("Error decoding config.json file:", err)
        return
    }

    // initial variables (linked to config.json)
    url := config.Planning[os.Args[1]]
    startDate := os.Args[2]
	endDate := os.Args[3]
    filename := config.Settings["filename"]
    calendarsFolder := config.Settings["calendarsFolder"]
    picturesFolder := config.Settings["picturesFolder"]
    mainColor := config.Settings["mainColor"]
    scdColor := config.Settings["scdColor"]
    textColor := config.Settings["textColor"]
    
	var newFilename string
    var newJSONname string

    newFilename, err = getICS(url, calendarsFolder, filename)
    if err != nil {
        fmt.Println("Error while retrieving ICS file:", err)
        return
        }

	newJSONname, err = getEvents(newFilename, calendarsFolder, startDate, endDate)
	if err != nil {
		fmt.Println("Error filling JSON file:", err)
		return
    	}
    err = filterEventsByDateRange(newJSONname, startDate, endDate)
    if err != nil {
        fmt.Println("Error filtering events by date:", err)
        return
    }
        
    err = CalendarGeneration(picturesFolder, newJSONname, mainColor, textColor, scdColor, newFilename)
    if err != nil {
        fmt.Println("Error generating calendar:", err)
        return
    }
}