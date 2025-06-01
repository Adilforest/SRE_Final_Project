package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BikeType string

const (
	RoadBike     BikeType = "road"
	MountainBike BikeType = "mountain"
	HybridBike   BikeType = "hybrid"
	ElectricBike BikeType = "electric"
)

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Price       float64            `bson:"price" json:"price"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Type        BikeType           `bson:"type" json:"type"`
	Brand       string             `bson:"brand" json:"brand"`
	Size        string             `bson:"size" json:"size"`
	Color       string             `bson:"color" json:"color"`
	Weight      float64            `bson:"weight" json:"weight"`
	Rating      float64            `bson:"rating" json:"rating"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	Features    []Feature          `bson:"features" json:"features"`
}

type Feature struct {
	Name  string `bson:"name" json:"name"`
	Value string `bson:"value" json:"value"`
}

type ProductFilter struct {
	Type      BikeType
	MinPrice  float64
	MaxPrice  float64
	Brands    []string
	Sizes     []string
	Search    string
	SortBy    string // "price", "rating", "newest"
	SortOrder int    // 1 (asc), -1 (desc)
}