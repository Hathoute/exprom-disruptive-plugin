import React, {ChangeEvent, PureComponent} from 'react';
import {LegacyForms} from '@grafana/ui';
import {DataSourcePluginOptionsEditorProps} from '@grafana/data';
import {MyDataSourceOptions} from '../types';

const { FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

interface CfgFormFieldProps {
  label: string,
  field: string,
  value: any
}

export class ConfigEditor extends PureComponent<Props, State> {

  onFieldChange = (event: ChangeEvent<HTMLInputElement>, field: string, isSecret?: boolean) => {
    const { onOptionsChange, options } = this.props;

    const parentFieldName = isSecret ? "secureJsonData" : "jsonData";
    const data = {
      ...options[parentFieldName],
      [field]: event.target.value,
    };
    onOptionsChange({ ...options, [parentFieldName]: data });
  };

  CfgFormField = (props: CfgFormFieldProps) => (
      <div className="gf-form">
        <FormField
            label={props.label}
            labelWidth={15}
            inputWidth={70}
            onChange={e => this.onFieldChange(e, props.field)}
            value={props.value || ''}
            placeholder="json field returned to frontend"
        />
      </div>
  )

  render() {
    const { options } = this.props;
    const { jsonData } = options;

    return (
      <div className="gf-form-group">
        <this.CfgFormField label="MongoDB Connection URL" field="mongodbUrl" value={jsonData.mongodbUrl}/>
      </div>
    );
  }
}
