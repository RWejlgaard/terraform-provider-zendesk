package zendesk

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	client "github.com/nukosuke/go-zendesk/zendesk"
)

// https://developer.zendesk.com/rest_api/docs/support/triggers
func resourceZendeskTrigger() *schema.Resource {
	return &schema.Resource{
		Create: func(d *schema.ResourceData, i interface{}) error {
			zd := i.(client.TriggerAPI)
			return createTrigger(d, zd)
		},
		Read: func(d *schema.ResourceData, i interface{}) error {
			zd := i.(client.TriggerAPI)
			return readTrigger(d, zd)
		},
		Update: func(d *schema.ResourceData, i interface{}) error {
			zd := i.(client.TriggerAPI)
			return updateTrigger(d, zd)
		},
		Delete: func(d *schema.ResourceData, i interface{}) error {
			zd := i.(client.TriggerAPI)
			return deleteTrigger(d, zd)
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"position": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			// Both the "all" and "any" parameter are optional, but at least one of them must be supplied
			"all": triggerConditionSchema(),
			"any": triggerConditionSchema(),
			"action": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func marshalTrigger(trigger client.Trigger, d identifiableGetterSetter) error {
	fields := map[string]interface{}{
		"title":       trigger.Title,
		"active":      trigger.Active,
		"position":    trigger.Position,
		"description": trigger.Description,
	}

	var alls []map[string]interface{}
	for _, v := range trigger.Conditions.All {
		m := map[string]interface{}{
			"field":    v.Field,
			"operator": v.Operator,
			"value":    v.Value,
		}
		alls = append(alls, m)
	}
	fields["all"] = alls

	var anys []map[string]interface{}
	for _, v := range trigger.Conditions.Any {
		m := map[string]interface{}{
			"field":    v.Field,
			"operator": v.Operator,
			"value":    v.Value,
		}
		anys = append(anys, m)
	}
	fields["any"] = anys

	var actions []map[string]interface{}
	for _, v := range trigger.Actions {
		m := map[string]interface{}{
			"field": v.Field,
			"value": v.Value,
		}
		actions = append(actions, m)
	}
	fields["actions"] = actions

	return setSchemaFields(d, fields)
}

func unmarshalTrigger(d identifiableGetterSetter) (client.Trigger, error) {
	trg := client.Trigger{}

	if v := d.Id(); v != "" {
		id, err := atoi64(v)
		if err != nil {
			return trg, fmt.Errorf("could not parse trigger id %s: %v", v, err)
		}
		trg.ID = id
	}

	if v, ok := d.GetOk("title"); ok {
		trg.Title = v.(string)
	}

	if v, ok := d.GetOk("active"); ok {
		trg.Active = v.(bool)
	}

	if v, ok := d.GetOk("description"); ok {
		trg.Description = v.(string)
	}

	if v, ok := d.GetOk("all"); ok {
		allConditions := v.(*schema.Set).List()
		conditions := []client.TriggerCondition{}
		for _, c := range allConditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				return trg, fmt.Errorf("could not parse 'all' conditions for trigger %v", trg)
			}
			conditions = append(conditions, client.TriggerCondition{
				Field:    condition["field"].(string),
				Operator: condition["operator"].(string),
				Value:    condition["value"].(string),
			})
		}
		trg.Conditions.All = conditions
	}

	if v, ok := d.GetOk("any"); ok {
		anyConditions := v.(*schema.Set).List()
		conditions := []client.TriggerCondition{}
		for _, c := range anyConditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				return trg, fmt.Errorf("could not parse 'any' conditions for trigger %v", trg)
			}
			conditions = append(conditions, client.TriggerCondition{
				Field:    condition["field"].(string),
				Operator: condition["operator"].(string),
				Value:    condition["value"].(string),
			})
		}
		trg.Conditions.Any = conditions
	}

	if v, ok := d.GetOk("actions"); ok {
		triggerActions := v.(*schema.Set).List()
		actions := []client.TriggerAction{}
		for _, a := range triggerActions {
			action, ok := a.(map[string]interface{})
			if !ok {
				return trg, fmt.Errorf("could not parse actions for trigger %v", trg)
			}
			actions = append(actions, client.TriggerAction{
				Field: action["field"].(string),
				Value: action["value"], // can take string or []string. So, let it be.
			})
		}
		trg.Actions = actions
	}

	return trg, nil
}

func createTrigger(d identifiableGetterSetter, zd client.TriggerAPI) error {
	trg, err := unmarshalTrigger(d)
	if err != nil {
		return err
	}

	trg, err = zd.CreateTrigger(trg)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", trg.ID))
	return marshalTrigger(trg, d)
}

func readTrigger(d *schema.ResourceData, zd client.TriggerAPI) error {
	return nil
}

func updateTrigger(d identifiableGetterSetter, zd client.TriggerAPI) error {
	return nil
}

func deleteTrigger(d identifiable, zd client.TriggerAPI) error {
	id, err := atoi64(d.Id())
	if err != nil {
		return err
	}

	return zd.DeleteTrigger(id)
}

func triggerConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeSet,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field": {
					Type:     schema.TypeString,
					Required: true,
				},
				"operator": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
		Optional: true,
	}
}
