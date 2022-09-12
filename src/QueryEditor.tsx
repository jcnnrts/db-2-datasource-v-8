import defaults from 'lodash/defaults';

import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './DataSource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

import AceEditor from "react-ace";
import "ace-builds/src-min-noconflict/ext-language_tools";
import "ace-builds/src-noconflict/mode-mysql";
import "ace-builds/src-noconflict/theme-terminal";


type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {

  onQueryChange = (newValue:String) => {
    const { onChange, query} = this.props;
    onChange({ ...query, queryText: newValue as any});

  };

  onQueryBlur = () => {
    const {onRunQuery} = this.props;
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText } = query;

    return (
      <div className="gf-form">

        <AceEditor
          placeholder=""
          mode="mysql"
          theme="terminal"
          name="qEditor"
          onChange={this.onQueryChange}
          onBlur={this.onQueryBlur}
          fontSize={14}
          height="200px"
          width="100%"
          showPrintMargin={true}
          showGutter={true}
          highlightActiveLine={true}
          value={queryText}
          setOptions={{
            enableBasicAutocompletion: true,
            enableLiveAutocompletion: true,
            enableSnippets: false,
            showLineNumbers: true,
            tabSize: 2,
          }}
        />

      </div>
    );
  }
}
