package models

import "time"

type Client struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type Employee struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Position  string    `json:"position"`
	ManagerID *int64    `json:"managerId,omitempty"`
	MentorID  *int64    `json:"mentorId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Campaign struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	StartDate  time.Time `json:"startDate"`
	FinishDate time.Time `json:"finishDate"`
	ClientID   int64     `json:"clientId"`
	ManagerID  int64     `json:"managerId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type AdPlatform struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CampaignPlatform struct {
	CampaignID  int64 `json:"campaignId"`
	PlatformID  int64 `json:"platformId"`
	BudgetCents int64 `json:"budgetCents"`
}

type AdSet struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	TargetAge     *string   `json:"targetAge,omitempty"`
	TargetGender  *string   `json:"targetGender,omitempty"`
	TargetCountry *string   `json:"targetCountry,omitempty"`
	CampaignID    int64     `json:"campaignId"`
	CreatedAt     time.Time `json:"createdAt"`
}

type MediaAsset struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	FilePath     string    `json:"filePath"`
	CreationDate time.Time `json:"creationDate"`
}

type Video struct {
	MediaAssetID int64 `json:"mediaAssetId"`
	Duration     int   `json:"duration"`
}

type Image struct {
	MediaAssetID int64  `json:"mediaAssetId"`
	Resolution   string `json:"resolution"`
}

type AdText struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type Ad struct {
	ID           int64     `json:"id"`
	AdSetID      int64     `json:"adSetId"`
	MediaAssetID int64     `json:"mediaAssetId"`
	AdTextID     int64     `json:"adTextId"`
	CreatedAt    time.Time `json:"createdAt"`
}
