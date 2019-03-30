/**
 * Copyright (c) 2017 eBay Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package goovn

import (
	"github.com/ebay/libovsdb"
)

type DHCPOptions struct {
	UUID       string
	CIDR       string
	Options    map[interface{}]interface{}
	ExternalID map[interface{}]interface{}
}

func (odbi *ovnDBImp) rowToDHCPOptions(uuid string) *DHCPOptions {
	cacheDHCPOptions, ok := odbi.cache[tableDHCPOptions][uuid]
	if !ok {
		return nil
	}

	dhcp := &DHCPOptions{
		UUID:       uuid,
		CIDR:       cacheDHCPOptions.Fields["cidr"].(string),
		Options:    cacheDHCPOptions.Fields["options"].(libovsdb.OvsMap).GoMap,
		ExternalID: cacheDHCPOptions.Fields["external_ids"].(libovsdb.OvsMap).GoMap,
	}

	return dhcp
}

func newDHCPRow(cidr string, options map[string]string, external_ids map[string]string) (OVNRow, error) {
	row := make(OVNRow)

	if len(cidr) > 0 {
		row["cidr"] = cidr
	}

	if options != nil {
		oMap, err := libovsdb.NewOvsMap(options)
		if err != nil {
			return nil, err
		}
		row["options"] = oMap
	}

	if external_ids != nil {
		oMap, err := libovsdb.NewOvsMap(external_ids)
		if err != nil {
			return nil, err
		}
		row["external_ids"] = oMap
	}

	return row, nil
}

func (odbi *ovnDBImp) dhcpOptionsAddImp(cidr string, options map[string]string, external_ids map[string]string) (*OvnCommand, error) {
	namedUUID, err := newRowUUID()
	if err != nil {
		return nil, err
	}

	row, err := newDHCPRow(cidr, options, external_ids)
	if err != nil {
		return nil, err
	}

	insertOp := libovsdb.Operation{
		Op:       opInsert,
		Table:    tableDHCPOptions,
		Row:      row,
		UUIDName: namedUUID,
	}

	operations := []libovsdb.Operation{insertOp}
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovnDBImp) dhcpOptionsSetImp(cidr string, options map[string]string, external_ids map[string]string) (*OvnCommand, error) {

	row, err := newDHCPRow(cidr, nil, external_ids)
	if err != nil {
		return nil, err
	}

	dhcpUUID := odbi.getRowUUID(tableDHCPOptions, row)
	if len(dhcpUUID) == 0 {
		return nil, ErrorNotFound
	}

	mutatemap, _ := libovsdb.NewOvsMap(options)
	mutation := libovsdb.NewMutation("options", opInsert, mutatemap)
	condition := libovsdb.NewCondition("_uuid", "==", libovsdb.UUID{dhcpUUID})

	mutateOp := libovsdb.Operation{
		Op:        opMutate,
		Table:     tableDHCPOptions,
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}

	operations := []libovsdb.Operation{mutateOp}
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovnDBImp) dhcpOptionsDelImp(uuid string) (*OvnCommand, error) {
	condition := libovsdb.NewCondition("_uuid", "==", libovsdb.UUID{uuid})
	deleteOp := libovsdb.Operation{
		Op:    opDelete,
		Table: tableDHCPOptions,
		Where: []interface{}{condition},
	}
	operations := []libovsdb.Operation{deleteOp}
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

// List all dhcp options
func (odbi *ovnDBImp) dhcpOptionsListImp() ([]*DHCPOptions, error) {
	var listDHCP []*DHCPOptions

	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheDHCPOptions, ok := odbi.cache[tableDHCPOptions]
	if !ok {
		return nil, ErrorSchema
	}

	for uuid, _ := range cacheDHCPOptions {
		listDHCP = append(listDHCP, odbi.rowToDHCPOptions(uuid))
	}
	return listDHCP, nil
}
