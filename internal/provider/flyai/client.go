package flyai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ariadne/internal/model"
)

type Client struct {
	Command string
}

func NewClient() Client {
	return Client{
		Command: "flyai",
	}
}

type hotelSearchResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ItemList []hotelItem `json:"itemList"`
	} `json:"data"`
}

type hotelItem struct {
	Address        string  `json:"address"`
	BrandName      *string `json:"brandName"`
	DecorationTime *string `json:"decorationTime"`
	DetailURL      string  `json:"detailUrl"`
	InterestsPOI   string  `json:"interestsPoi"`
	Latitude       string  `json:"latitude"`
	Longitude      string  `json:"longitude"`
	MainPic        string  `json:"mainPic"`
	Name           string  `json:"name"`
	Price          string  `json:"price"`
	ShID           string  `json:"shId"`
	Star           string  `json:"star"`
}

func (c Client) SearchHotels(
	ctx context.Context,
	destination string,
	poiName string,
	checkInDate string,
	checkOutDate string,
	maxPrice int,
) ([]model.HotelOffer, error) {
	if destination == "" {
		return nil, errors.New("destination is empty")
	}

	if checkInDate == "" || checkOutDate == "" {
		return nil, errors.New("hotel check-in date or check-out date is empty")
	}

	args := []string{
		"search-hotel",
		"--dest-name", destination,
		"--check-in-date", checkInDate,
		"--check-out-date", checkOutDate,
		"--sort", "distance_asc",
	}

	if poiName != "" {
		args = append(args, "--poi-name", poiName)
	}

	if maxPrice > 0 {
		args = append(args, "--max-price", strconv.Itoa(maxPrice))
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, c.Command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	rawOutput := strings.TrimSpace(stdout.String())
	if rawOutput == "" {
		if err != nil {
			return nil, errors.New("flyai command failed: " + err.Error() + "; stderr: " + stderr.String())
		}
		return nil, errors.New("flyai returned empty output")
	}

	jsonText := extractFirstJSON(rawOutput)

	var resp hotelSearchResponse
	if jsonErr := json.Unmarshal([]byte(jsonText), &resp); jsonErr != nil {
		return nil, jsonErr
	}

	if resp.Status != 0 {
		return nil, errors.New("flyai hotel search failed: " + resp.Message)
	}

	nights := calculateNights(checkInDate, checkOutDate)

	offers := make([]model.HotelOffer, 0)

	for _, item := range resp.Data.ItemList {
		price := parseCNYPrice(item.Price)
		lat := parseFloat(item.Latitude)
		lng := parseFloat(item.Longitude)

		offers = append(offers, model.HotelOffer{
			Provider:      "fliggy",
			Name:          item.Name,
			Address:       item.Address,
			Star:          item.Star,
			PricePerNight: price,
			TotalPrice:    price * nights,
			Nights:        nights,
			BookingLink:   item.DetailURL,
			ImageURL:       item.MainPic,
			NearbyPOI:     item.InterestsPOI,
			Lat:           lat,
			Lng:           lng,
			DataSource:    "flyai_fliggy",
			Status:        "ok",
			Message:       "query ok",
		})
	}

	return offers, nil
}

type trainSearchResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ItemList []trainItem `json:"itemList"`
	} `json:"data"`
}

type trainItem struct {
	Journeys      []trainJourney `json:"journeys"`
	JumpURL       string         `json:"jumpUrl"`
	Price         string         `json:"price"`
	TotalDuration string         `json:"totalDuration"`
}

type trainJourney struct {
	JourneyType      string         `json:"journeyType"`
	Segments         []trainSegment `json:"segments"`
	TotalDuration    string         `json:"totalDuration"`
	TransferDuration string         `json:"transferDuration"`
}

type trainSegment struct {
	ArrCityName            string `json:"arrCityName"`
	ArrDateTime            string `json:"arrDateTime"`
	ArrStationName         string `json:"arrStationName"`
	DepCityName            string `json:"depCityName"`
	DepDateTime            string `json:"depDateTime"`
	DepStationName         string `json:"depStationName"`
	Duration               string `json:"duration"`
	MarketingTransportName string `json:"marketingTransportName"`
	MarketingTransportNo   string `json:"marketingTransportNo"`
	SeatClassName          string `json:"seatClassName"`
	TransportType          string `json:"transportType"`
}

func (c Client) SearchTrains(
	ctx context.Context,
	origin string,
	destination string,
	depDate string,
	seatClassName string,
	maxPrice int,
) ([]model.TrainOffer, error) {
	if origin == "" {
		return nil, errors.New("origin is empty")
	}

	if destination == "" {
		return nil, errors.New("destination is empty")
	}

	if depDate == "" {
		return nil, errors.New("train departure date is empty")
	}

	args := []string{
		"search-train",
		"--origin", origin,
		"--destination", destination,
		"--dep-date", depDate,
		"--sort-type", "3",
	}

	if seatClassName != "" {
		args = append(args, "--seat-class-name", seatClassName)
	}

	if maxPrice > 0 {
		args = append(args, "--max-price", strconv.Itoa(maxPrice))
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, c.Command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	rawOutput := strings.TrimSpace(stdout.String())
	if rawOutput == "" {
		if err != nil {
			return nil, errors.New("flyai train command failed: " + err.Error() + "; stderr: " + stderr.String())
		}
		return nil, errors.New("flyai train returned empty output")
	}

	jsonText := extractFirstJSON(rawOutput)

	var resp trainSearchResponse
	if jsonErr := json.Unmarshal([]byte(jsonText), &resp); jsonErr != nil {
		return nil, jsonErr
	}

	if resp.Status != 0 {
		return nil, errors.New("flyai train search failed: " + resp.Message)
	}

	offers := make([]model.TrainOffer, 0)

	for _, item := range resp.Data.ItemList {
		offer := model.TrainOffer{
			Provider:             "fliggy",
			Price:                parseCNYPrice(item.Price),
			TotalDurationMinutes: parseInt(item.TotalDuration),
			BookingLink:          item.JumpURL,
			DataSource:           "flyai_fliggy",
			Status:               "ok",
			Message:              "query ok",
		}

		if len(item.Journeys) > 0 {
			journey := item.Journeys[0]
			offer.JourneyType = journey.JourneyType

			segments := make([]model.TrainSegment, 0)

			for _, segment := range journey.Segments {
				segments = append(segments, model.TrainSegment{
					TrainNo:         segment.MarketingTransportNo,
					TrainType:       segment.MarketingTransportName,
					SeatClassName:   segment.SeatClassName,
					DepCityName:     segment.DepCityName,
					DepStationName:  segment.DepStationName,
					DepDateTime:     segment.DepDateTime,
					ArrCityName:     segment.ArrCityName,
					ArrStationName:  segment.ArrStationName,
					ArrDateTime:     segment.ArrDateTime,
					DurationMinutes: parseInt(segment.Duration),
				})
			}

			offer.Segments = segments
		}

		offers = append(offers, offer)
	}

	return offers, nil
}

func parseInt(text string) int {
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0
	}

	return value
}

func extractFirstJSON(text string) string {
	text = strings.TrimSpace(text)

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start >= 0 && end > start {
		return text[start : end+1]
	}

	return text
}

func parseCNYPrice(priceText string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(priceText)

	if match == "" {
		return 0
	}

	price, err := strconv.Atoi(match)
	if err != nil {
		return 0
	}

	return price
}

func parseFloat(text string) float64 {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0
	}

	return value
}

func calculateNights(checkInDate string, checkOutDate string) int {
	in, err1 := time.Parse("2006-01-02", checkInDate)
	out, err2 := time.Parse("2006-01-02", checkOutDate)

	if err1 != nil || err2 != nil {
		return 1
	}

	nights := int(out.Sub(in).Hours() / 24)
	if nights <= 0 {
		return 1
	}

	return nights
}