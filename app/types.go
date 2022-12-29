package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type Story struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Content   string             `json:"content,omitempty" bson:"content,omitempty"`
	Title     string             `json:"title,omitempty" bson:"title,omitempty"`
	Tokens    []JpToken          `json:"tokens,omitempty" bson:"tokens,omitempty"`
	Sentences []Sentence
}

type JpToken struct {
	Surface          string `json:"surface,omitempty" bson:"surface,omitempty"`
	POS              string `json:"pos,omitempty" bson:"pos"`
	POS_1            string `json:"pos1,omitempty" bson:"pos1"`
	POS_2            string `json:"pos2,omitempty" bson:"pos2"`
	POS_3            string `json:"pos3,omitempty" bson:"pos3"`
	InflectionalType string `json:"inflectionalType,omitempty" bson:"inflectionalType"`
	InflectionalForm string `json:"inflectionalForm,omitempty" bson:"inflectionalForm"`
	BaseForm         string `json:"baseForm,omitempty" bson:"baseForm"`
	Reading          string `json:"reading,omitempty" bson:"reading"`
	Pronunciation    string `json:"pronunciation,omitempty" bson:"pronunciation"`
}

type Sentence struct {
	Words          []Word
	EndPunctuation string
}

type Word struct {
	Text       string             `json:"text,omitempty" bson:"text,omitempty"`
	Definition primitive.ObjectID `json:"definition,omitempty" bson:"definition,omitempty"`
	Form       string             `json:"form,omitempty" bson:"form,omitempty"` // e.g. 'ました' for a verb
}
