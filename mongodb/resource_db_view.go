package mongodb

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatabaseView() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseViewCreate,
		ReadContext:   resourceDatabaseViewRead,
		UpdateContext: resourceDatabaseViewUpdate,
		DeleteContext: resourceDatabaseViewDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"view_on": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pipeline": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDatabaseViewCreate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var config = i.(*MongoDatabaseConfiguration)
	client, connectionError := MongoClientInit(config)
	if connectionError != nil {
		return diag.Errorf("Error connecting to database : %s ", connectionError)
	}

	var view = data.Get("name").(string)
	var database = data.Get("database").(string)
	var collection = data.Get("view_on").(string)
	var pipeline = data.Get("pipeline").(string)

	err := createView(client, view, collection, pipeline, database)

	if err != nil {
		return diag.Errorf("Could not create the view: %s ", err)
	}
	str := database + "." + view
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	data.SetId(encoded)
	return resourceDatabaseViewRead(ctx, data, i)
}

func resourceDatabaseViewDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var config = i.(*MongoDatabaseConfiguration)
	client, connectionError := MongoClientInit(config)
	if connectionError != nil {
		return diag.Errorf("Error connecting to database : %s", connectionError)
	}
	var stateId = data.State().ID
	name, database, err := resourceDatabaseViewParseId(stateId)

	if err != nil {
		return diag.Errorf("%s", err)
	}

	err = client.Database(database).Collection(name).Drop(context.TODO())

	if err != nil {
		return diag.Errorf("%s", err)
	}
	data.SetId("")

	return nil
}

func resourceDatabaseViewParseId(id string) (string, string, error) {
	result, errEncoding := base64.StdEncoding.DecodeString(id)

	if errEncoding != nil {
		return "", "", fmt.Errorf("unexpected format of ID Error : %s", errEncoding)
	}
	parts := strings.SplitN(string(result), ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1.attribute2", id)
	}

	database := parts[0]
	view := parts[1]

	return view, database, nil
}

func resourceDatabaseViewUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	resourceDatabaseViewDelete(ctx, data, i)
	return resourceDatabaseViewCreate(ctx, data, i)
}

func resourceDatabaseViewRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var config = i.(*MongoDatabaseConfiguration)
	client, connectionError := MongoClientInit(config)
	if connectionError != nil {
		return diag.Errorf("Error connecting to database : %s ", connectionError)
	}
	stateID := data.State().ID
	name, database, err := resourceDatabaseViewParseId(stateID)
	if err != nil {
		return diag.Errorf("%s", err)
	}
	_, decodeError := getView(client, name, database)
	if decodeError != nil {
		return diag.Errorf("Error decoding view: %s ", decodeError)
	}

	dataSetError := data.Set("name", name)
	if dataSetError != nil {
		return diag.Errorf("error setting role: %s ", dataSetError)
	}
	dataSetError = data.Set("database", database)
	if dataSetError != nil {
		return diag.Errorf("error setting database: %s ", dataSetError)
	}
	data.SetId(stateID)
	return nil
}
