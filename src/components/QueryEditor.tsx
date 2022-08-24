import {defaults} from 'lodash';

import React, {useEffect, useState} from 'react';
import {ActionMeta, LegacyForms, Select} from '@grafana/ui';
import {MetricFindValue, QueryEditorProps, SelectableValue} from '@grafana/data';
import {DataSource} from '../datasource';
import {defaultQuery, MyDataSourceOptions, MyQuery} from '../types';

const { Switch } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor(props: Props) {

    const onChange = (query: MyQuery) => {
        props.onChange(query);
        props.onRunQuery();
    }

    const createSelect = (label: string, options: SelectableValue[], value: any,
                          onChange: (e: SelectableValue, actionMeta: ActionMeta) => void) => (
        <>
            <span className="gf-form-label width-10">{label}</span>
            <Select
                options={options}
                value={value}
                onChange={onChange}
                allowCustomValue={false}
                closeMenuOnSelect={true}
                isClearable={false}
                isMulti={true}
            />
        </>
    )

    const toSelectableValue = (entity: MetricFindValue): SelectableValue<string> => {
        return {
            label: entity.text,
            value: entity.value as string
        }
    };

    const setParam = (p: string, v: any) => {
        setQuery({
            ...query,
            parameters: {
                ...query.parameters,
                [p]: v
            }
        });
    }

    const onProjectsChange = (e: SelectableValue<string | number>) => {
        let projects = "" + e.value;
        if(e instanceof Array) {
            projects = e.map(x => x.value).join(",")
        }
        setParam("projects", projects);
    }

    const onDevicesChange = (e: SelectableValue<string | number>) => {
        let devices = "" + e.value;
        if(e instanceof Array) {
            devices = e.map(x => x.value).join(",")
        }
        setParam("devices", devices);
    }

    const [query, setQuery] = useState<MyQuery>(defaults(props.query, defaultQuery));

    const [allProjects, setAllProjects] = useState<MetricFindValue[]>([]);
    const [allDevices, setAllDevices] = useState<MetricFindValue[]>([]);

    const selectedProjects = query.parameters["projects"];
    const selectedProjectsArray = selectedProjects.split(",");
    const selectedDevices = query.parameters["devices"];
    const selectedDevicesArray = selectedDevices.split(",");

    console.log("allProjects", allProjects)
    console.log("allDevices", allDevices)

    useEffect(() => {
        props.datasource.metricFindQuery({entity: "Projects"}).then(r => setAllProjects(r));
    }, [props.datasource]);

    useEffect(() => {
        props.datasource.metricFindQuery({entity: "Devices", projects: selectedProjects ?? "-1"})
            .then(r => setAllDevices(r));
    }, [props.datasource, selectedProjects])

    useEffect(() => {
        if(query.entity !== "Events") {
            setQuery({...query, entity: "Events"});
            return;   // onChange will get executed next time (dependency on query)
        }

        if(query.parameters["filter"] !== "devices") {
            setParam("filter", "devices");  // We are displaying events, so enforce devices as filter.
            return;
        }

        onChange(query)

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [query]);

    return (
        <div>
            <div className="gf-form">
                <span className="gf-form-label width-10">FROM</span>
                {createSelect("Projects",
                    allProjects.map(toSelectableValue),
                    allProjects.filter(x => selectedProjectsArray.includes(x.value as string)),
                    onProjectsChange
                )}
            </div>

            <div className="gf-form">
                <span className="gf-form-label width-10">SELECT</span>
                {createSelect("Devices",
                    allDevices.map(toSelectableValue),
                    allDevices.filter(x => selectedDevicesArray.includes(x.value as string)),
                    onDevicesChange
                )}
            </div>

            <div className="gf-form">
                <Switch checked={query.withStreaming}
                        label="Enable streaming (v8+)"
                        onChange={e => setQuery({
                            ...query,
                            withStreaming: e.currentTarget.checked
                        })} />
            </div>
        </div>
    );
}
