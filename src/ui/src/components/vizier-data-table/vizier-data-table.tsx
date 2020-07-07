import { Table } from 'common/vizier-grpc-client';
import {
  CellAlignment, ColumnProps, DataTable, SortState,
} from 'components/data-table';
import { JSONData } from 'components/format-data/format-data';
import * as React from 'react';
import { DataType, Relation, SemanticType } from 'types/generated/vizier_pb';
import noop from 'utils/noop';
import { dataFromProto } from 'utils/result-data-utils';
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';
import { parseRows } from './parsers';
import { vizierCellRenderer } from './renderers';
import { getSortFunc } from './sort-funcs';

const DataAlignmentMap = new Map<DataType, CellAlignment>(
  [
    [DataType.BOOLEAN, 'center'],
    [DataType.INT64, 'end'],
    [DataType.UINT128, 'start'],
    [DataType.FLOAT64, 'end'],
    [DataType.STRING, 'start'],
    [DataType.TIME64NS, 'start'],
    [DataType.DURATION64NS, 'start'],
  ],
);

// For certain semantic types, override the column width ratio based off of the rendering
// we expect to do for that semantic type.
const SemanticTypeWidthOverrideMap = new Map<SemanticType, number>(
  [
    [SemanticType.ST_QUANTILES, 40],
  ],
);

interface VizierDataTableProps {
  table: Table;
  prettyRender?: boolean;
  expandable?: boolean;
  expandedRenderer?: (rowIndex: number) => JSX.Element;
  // TODO(michelle/zasgar/nserrino): Remove this.
  clusterName?: string;
  onRowSelectionChanged?: (row: any) => void;
}

interface TypedColumnProps extends ColumnProps {
  type: DataType;
  semanticType: SemanticType;
}

export const VizierDataTable = (props: VizierDataTableProps) => {
  const {
    table, prettyRender = false, expandable = false, expandedRenderer,
    clusterName = null,
    onRowSelectionChanged = noop,
  } = props;
  const [rows, setRows] = React.useState([]);
  const [selectedRow, setSelectedRow] = React.useState(-1);
  const [columnsMap, setColumnsMap] = React.useState<Map<string, TypedColumnProps>>(
    new Map<string, TypedColumnProps>());

  React.useEffect(() => {
    // Map containing the display information for the column.
    const newMap = new Map<string, TypedColumnProps>();
    // Semantic type map which is used by parsers to parse certain semantic types with
    // special handling.
    const semanticTypeMap = new Map<string, SemanticType>();

    table.relation.getColumnsList().forEach((col) => {
      const name = col.getColumnName();
      semanticTypeMap.set(name, col.getColumnSemanticType());
    });

    const rawRows = dataFromProto(table.relation, table.data);
    const parsedRows = parseRows(semanticTypeMap, rawRows);

    table.relation.getColumnsList().forEach((col) => {
      const name = col.getColumnName();
      const columnProp: TypedColumnProps = {
        type: col.getColumnType(),
        semanticType: col.getColumnSemanticType(),
        dataKey: col.getColumnName(),
        label: col.getColumnName(),
        align: DataAlignmentMap.get(col.getColumnType()) || 'start',
        cellRenderer: vizierCellRenderer(prettyRender, col, clusterName, parsedRows),
      };
      if (SemanticTypeWidthOverrideMap.has(columnProp.semanticType)) {
        columnProp.width = SemanticTypeWidthOverrideMap.get(columnProp.semanticType);
      }

      newMap.set(name, columnProp);
      semanticTypeMap.set(name, col.getColumnSemanticType());
    });

    setRows(parsedRows);
    setColumnsMap(newMap);
  }, [table.relation, table.data, clusterName, prettyRender]);

  const rowGetter = React.useCallback(
    (i) => rows[i],
    [rows],
  );

  const onSort = (sortState: SortState) => {
    const column = columnsMap.get(sortState.dataKey);
    setRows(rows.sort(getSortFunc(sortState.dataKey, column.type, column.semanticType, sortState.direction)));
    setSelectedRow(-1);
    onRowSelectionChanged(null);
  };

  const onRowSelect = React.useCallback((rowIndex) => {
    if (rowIndex === selectedRow) {
      rowIndex = -1;
    }
    setSelectedRow(rowIndex);
    onRowSelectionChanged(rows[rowIndex]);
  }, [rows, selectedRow, onRowSelectionChanged]);

  if (rows.length === 0) {
    return null;
  }

  return (
    <DataTable
      rowGetter={rowGetter}
      rowCount={rows.length}
      columns={[...columnsMap.values()]}
      compact
      onSort={onSort}
      onRowClick={onRowSelect}
      highlightedRow={selectedRow}
      expandable={expandable}
      expandedRenderer={expandedRenderer}
    />
  );
};

const useStyles = makeStyles((theme: Theme) => createStyles({
  root: {
    display: 'flex',
    flexDirection: 'row',
    height: '100%',
    position: 'relative',
  },
  table: {
    flex: 3,
  },
  details: {
    flex: 1,
    padding: theme.spacing(2),
    borderLeft: `solid 1px ${theme.palette.background.three}`,
    minWidth: 0,
    overflow: 'auto',
  },
  close: {
    position: 'absolute',
  },
}));

export const VizierDataTableWithDetails = (props: { table: Table }) => {
  const [details, setDetails] = React.useState(null);

  const classes = useStyles();
  return (
    <div className={classes.root}>
      <div className={classes.table}>
        <VizierDataTable expandable={false} table={props.table} onRowSelectionChanged={(row) => { setDetails(row); }} />
      </div>
      <VizierDataRowDetails className={classes.details} data={details} />
    </div>
  );
};

interface VizierDataRowDetailsProps {
  data?: any;
  className?: string;
}

const VizierDataRowDetails = (props: VizierDataRowDetailsProps) => {
  const { data, className } = props;
  if (!data) {
    return null;
  }
  return <JSONData className={className} data={data} multiline />;
};
