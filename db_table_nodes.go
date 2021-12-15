package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Node struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Nodename            string         `gorm:"column:nodename; index" json:"nodename"`
	NodeID              string         `gorm:"column:node_id; uniqueIndex; size:36" json:"node_id"`
	ClusterID           string         `gorm:"column:cluster_id; size:36; index" json:"cluster_id"`
	WarrantyEnd         time.Time      `gorm:"column:warranty_end; autoCreateTime" json:"warranty_end"`
	MaintenanceEnd      time.Time      `gorm:"column:maintenance_end; autoCreateTime" json:"maintenance_end"`
	Status              string         `gorm:"column:status" json:"status"`
	Role                string         `gorm:"column:role" json:"role"`
	ListenerPort        int            `gorm:"column:listener_port; default:1214" json:"listener_port"`
	Version             string         `gorm:"column:version" json:"version"`
	Collector           string         `gorm:"column:collector" json:"collector"`
	ConnectTo           string         `gorm:"column:connect_to" json:"connect_to"`
	LastComm            time.Time      `gorm:"column:last_comm" json:"last_comm"`
	TZ                  string         `gorm:"column:tz" json:"tz"`
	AssetEnv            string         `gorm:"column:asset_env" json:"asset_env"`
	NodeEnv             string         `gorm:"column:node_env; default:TST" json:"node_env"`
	MemBytes            int            `gorm:"column:mem_bytes" json:"mem_bytes"`
	MemBanks            int            `gorm:"column:mem_banks" json:"mem_banks"`
	MemSlots            int            `gorm:"column:mem_slots" json:"mem_slots"`
	OSVendor            string         `gorm:"column:os_vendor; index" json:"os_vendor"`
	OSName              string         `gorm:"column:os_name; index" json:"os_name"`
	OSKernel            string         `gorm:"column:os_kernel" json:"os_kernel"`
	OSRelease           string         `gorm:"column:os_release; index" json:"os_release"`
	OSArch              string         `gorm:"column:os_arch; index" json:"os_arch"`
	OSConcat            string         `gorm:"column:os_concat" json:"os_concat"`
	CPUFreq             int            `gorm:"column:cpu_freq" json:"cpu_freq"`
	CPUDies             int            `gorm:"column:cpu_dies" json:"cpu_dies"`
	CPUCores            int            `gorm:"column:cpu_cores" json:"cpu_cores"`
	CPUThreads          int            `gorm:"column:cpu_threads" json:"cpu_threads"`
	CPUModel            string         `gorm:"column:cpu_model" json:"cpu_model"`
	CPUVendor           string         `gorm:"column:cpu_vendor" json:"cpu_vendor"`
	Type                string         `gorm:"column:type" json:"type"`
	FQDN                string         `gorm:"column:fqdn" json:"fqdn"`
	TeamResponsible     string         `gorm:"column:team_responsible; index" json:"team_responsible"`
	TeamInteg           string         `gorm:"column:team_integ" json:"team_integ"`
	TeamSupport         string         `gorm:"column:team_support" json:"team_support"`
	App                 string         `gorm:"column:app" json:"app"`
	Serial              string         `gorm:"column:serial" json:"serial"`
	SPVersion           string         `gorm:"column:sp_version" json:"sp_version"`
	BIOSVersion         string         `gorm:"column:bios_version" json:"bios_version"`
	Manufacturer        string         `gorm:"column:manufacturer" json:"manufacturer"`
	Model               string         `gorm:"column:model; index" json:"model"`
	LocAddr             string         `gorm:"column:loc_addr" json:"loc_addr"`
	LocCity             string         `gorm:"column:loc_city" json:"loc_city"`
	LocZIP              string         `gorm:"column:loc_zip" json:"loc_zip"`
	LocRack             string         `gorm:"column:loc_rack" json:"loc_rack"`
	LocFloor            string         `gorm:"column:loc_floor" json:"loc_floor"`
	LocCountry          string         `gorm:"column:loc_country" json:"loc_country"`
	LocBuilding         string         `gorm:"column:loc_building" json:"loc_building"`
	LocRoom             string         `gorm:"column:loc_room" json:"loc_room"`
	PowerSupplyNb       int            `gorm:"column:power_supply_nb; default:0" json:"power_supply_nb"`
	PowerCabinet1       string         `gorm:"column:power_cabinet1" json:"power_cabinet1"`
	PowerCabinet2       string         `gorm:"column:power_cabinet2" json:"power_cabinet2"`
	PowerProtect        string         `gorm:"column:power_protect" json:"power_protect"`
	PowerProtectBreaker string         `gorm:"column:power_protect_breaker" json:"power_protect_breaker"`
	PowerBreaker1       string         `gorm:"column:power_breaker1" json:"power_breaker1"`
	PowerBreaker2       string         `gorm:"column:power_breaker2" json:"power_breaker2"`
	Updated             time.Time      `gorm:"column:updated" json:"updated"`
	Enclosure           string         `gorm:"column:enclosure" json:"enclosure"`
	EnclosureSlot       string         `gorm:"column:enclosureslot" json:"enclosureslot"`
	AssetName           string         `gorm:"column:assetname" json:"assetname"`
	SecZone             string         `gorm:"column:sec_zone" json:"sec_zone"`
	LastBoot            time.Time      `gorm:"column:last_boot" json:"last_boot"`
	ActionType          string         `gorm:"column:action_type; type:enum('push', 'pull'); default:pull" json:"action_type"`
	HVPool              string         `gorm:"column:hvpool" json:"hvpool"`
	HVVDC               string         `gorm:"column:hvvdc" json:"hvvdc"`
	HV                  string         `gorm:"column:hv" json:"hv"`
	HWObsWarnDate       time.Time      `gorm:"column:hw_obs_warn_date" json:"hw_obs_warn_date"`
	HWObsAlertDate      time.Time      `gorm:"column:hw_obs_alert_date" json:"hw_obs_alert_date"`
	OSObsWarnDate       time.Time      `gorm:"column:os_obs_warn_date" json:"os_obs_warn_date"`
	OSObsAlertDate      time.Time      `gorm:"column:os_obs_alert_date" json:"os_obs_alert_date"`
	Notifications       bool           `gorm:"column:notifications; default:true" json:"notifications"`
	SnoozeTill          time.Time      `gorm:"column:snooze_till" json:"snooze_till"`
	NodeFrozen          bool           `gorm:"column:node_frozen" json:"node_frozen"`
}

func init() {
	tables["nodes"] = newTable("nodes").SetEntry(Node{})
}

var (
	reUUID, _ = regexp.Compile("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
	reID, _   = regexp.Compile("[0-9]+")
)

func nodeCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if reUUID.MatchString(id) {
			n, err := getNodeByNodeID(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "node", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else if reID.MatchString(id) {
			n, err := getNodeByID(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "node", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			n, err := getNodeByName(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "node", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getNodeByNodeID(nodeID string) (Node, error) {
	data := make([]Node, 0)
	result := db.Where("node_id = ?", nodeID).Find(&data)
	if result.Error != nil {
		return Node{}, result.Error
	}
	if len(data) == 0 {
		return Node{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getNodeByName(name string) (Node, error) {
	data := make([]Node, 0)
	result := db.Where("nodename = ?", name).Find(&data)
	if result.Error != nil {
		return Node{}, result.Error
	}
	if len(data) == 0 {
		return Node{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getNodeByID(id string) (Node, error) {
	data := make([]Node, 0)
	result := db.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return Node{}, result.Error
	}
	if len(data) == 0 {
		return Node{}, fmt.Errorf("not found")
	}
	return data[0], nil
}
