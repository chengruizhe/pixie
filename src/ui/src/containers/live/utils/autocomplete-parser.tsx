interface TabStop {
  Index: number;
  Label?: string;
  Value?: string;
  CursorPosition: number; // -1 if cursor is not contained in this tabstop.
}

const tabStopRegex = /\${(.*?)}/g; // Match anything between braces: ${___}
const tabStopIndexRegex = /{[0-9]+/g; // Match any number following a brace: ${__
const tabStopLabelRegex = /:(.*?):/; // Match any text between two colons :__:
const tabStopValueRegex = /:([^:]*?)}/; // Match any text following a colon before a brace :__}

export function ParseFormatStringToTabStops(input: string) {
  const tabStops = [];

  const parsedTS = [...input.match(tabStopRegex)]; // Tokenize into tabstops.
  parsedTS.forEach((ts) => {
    const tabStop = {CursorPosition: -1} as TabStop;

    const idx = ts.match(tabStopIndexRegex)[0].substring(1); // Parse tabstop index. This is always defined.
    tabStop.Index = parseInt(idx, 10);

    const label = ts.match(tabStopLabelRegex); // Parse tabstop label.
    if (label !== null && label[1].length > 0) {
      tabStop.Label = label[1];
    }

    const value = ts.match(tabStopValueRegex); // Parse tabstop value.
    if (value !== null && value[1].length > 0) {
      tabStop.CursorPosition = value[1].indexOf('$0');
      if (tabStop.CursorPosition !== -1) { // Find cursor position.
        tabStop.Value = value[1].replace('$0', '');
      } else {
        tabStop.Value = value[1];
      }
    }

    tabStops.push(tabStop);
  });

  return tabStops;
}
