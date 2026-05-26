package main

import "time"

type agentSeed struct {
	ID             string
	DisplayName    string
	PhoneNumber    string
	WhatsAppNumber string
}

type propertySeed struct {
	ID          string
	Name        string
	Area        string
	Landmark    string
	Description string
	CreatedAt   time.Time
}

type propertyUnitTypeSeed struct {
	ID           string
	PropertyID   string
	Category     string
	Name         string
	Description  string
	Notes        string
	BedroomCount int
	HasParlour   bool
	BathroomType string
	KitchenType  string
	CreatedAt    time.Time
}

type agentOfferSeed struct {
	ID          string
	UnitTypeID  string
	AgentID     string
	Title       string
	Description string
	Notes       string
	PriceKobo   int
	Status      string
	CreatedAt   time.Time
}

var seedBaseTime = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)

var agents = []agentSeed{
	{
		ID:             "11111111-1111-4111-8111-111111111111",
		DisplayName:    "Tunde Campus Agent",
		PhoneNumber:    "+2348010000001",
		WhatsAppNumber: "+2348010000001",
	},
	{
		ID:             "22222222-2222-4222-8222-222222222222",
		DisplayName:    "Bisi Housing Connect",
		PhoneNumber:    "+2348010000002",
		WhatsAppNumber: "+2348010000002",
	},
}

var properties = []propertySeed{
	{
		ID:          "33333333-3333-4333-8333-333333333333",
		Name:        "Alice Lodge",
		Area:        "Obanla",
		Landmark:    "Near South Gate",
		Description: "Gated student lodge with multiple room categories and steady access from campus.",
		CreatedAt:   seedBaseTime.Add(2 * time.Hour),
	},
	{
		ID:          "44444444-4444-4444-8444-444444444444",
		Name:        "Blue Roof Apartments",
		Area:        "Aule",
		Landmark:    "Off Akure-Ilesha Road",
		Description: "Quiet compound with self-contained rooms and agent-distributed availability.",
		CreatedAt:   seedBaseTime.Add(time.Hour),
	},
	{
		ID:          "55555555-5555-4555-8555-555555555555",
		Name:        "South Gate Residence",
		Area:        "South Gate",
		Landmark:    "Behind mini market",
		Description: "Walkable accommodation cluster with mixed unit types for students comparing options.",
		CreatedAt:   seedBaseTime,
	},
}

var unitTypes = []propertyUnitTypeSeed{
	{"66666666-6666-4666-8666-666666666661", "33333333-3333-4333-8333-333333333333", "single_room", "Single Room", "Basic single room in the lodge compound.", "Upstairs unit near water tank. Shared bathroom is cleaned weekly.", 1, false, "shared", "shared", seedBaseTime},
	{"66666666-6666-4666-8666-666666666662", "33333333-3333-4333-8333-333333333333", "self_contained", "Self-contained", "Private room with toilet and bathroom.", "Top floor corner unit with better ventilation.", 1, false, "private", "private", seedBaseTime.Add(10 * time.Minute)},
	{"77777777-7777-4777-8777-777777777771", "44444444-4444-4444-8444-444444444444", "room_and_parlour", "Room and Parlour", "Larger unit suited for students who want extra space.", "Kitchen is outside but private. Gate closes by 10pm.", 1, true, "private", "private", seedBaseTime.Add(20 * time.Minute)},
	{"77777777-7777-4777-8777-777777777772", "44444444-4444-4444-8444-444444444444", "self_contained", "Self-contained", "Self-contained room in a quieter part of Aule.", "Ground floor. Landlord lives on site.", 1, false, "private", "private", seedBaseTime.Add(30 * time.Minute)},
	{"88888888-8888-4888-8888-888888888881", "55555555-5555-4555-8555-555555555555", "single_room", "Single Room", "Budget single room near South Gate.", "Shared kitchen gets busy in the morning.", 1, false, "shared", "shared", seedBaseTime.Add(40 * time.Minute)},
}

var agentOffers = []agentOfferSeed{
	{"99999999-9999-4999-8999-999999999991", "66666666-6666-4666-8666-666666666661", "11111111-1111-4111-8111-111111111111", "Obanla single room", "Agent has current access for inspection coordination.", "2 left. Inspection tomorrow only.", 18000000, "available", seedBaseTime},
	{"99999999-9999-4999-8999-999999999992", "66666666-6666-4666-8666-666666666662", "22222222-2222-4222-8222-222222222222", "Fresh self-contained", "Recently painted self-contained option in Alice Lodge.", "Landlord prefers students. Can negotiate for 2-year rent.", 35000000, "available", seedBaseTime.Add(10 * time.Minute)},
	{"99999999-9999-4999-8999-999999999993", "77777777-7777-4777-8777-777777777771", "11111111-1111-4111-8111-111111111111", "Room and parlour in Aule", "Spacious unit with flexible inspection timing.", "1 left. Agent is out of town until Friday.", 50000000, "available", seedBaseTime.Add(20 * time.Minute)},
	{"99999999-9999-4999-8999-999999999994", "77777777-7777-4777-8777-777777777772", "22222222-2222-4222-8222-222222222222", "Quiet Aule self-contained", "Availability needs reconfirmation before inspection.", "Paused while landlord decides on renovation.", 32000000, "paused", seedBaseTime.Add(30 * time.Minute)},
	{"99999999-9999-4999-8999-999999999995", "88888888-8888-4888-8888-888888888881", "11111111-1111-4111-8111-111111111111", "Budget South Gate room", "Lower-priced option near student movement routes.", "3 available. First come, no reservation.", 15000000, "available", seedBaseTime.Add(40 * time.Minute)},
}
