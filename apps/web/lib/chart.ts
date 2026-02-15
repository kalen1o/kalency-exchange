import { Candle } from "./api";

function escapeXml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll("\"", "&quot;")
    .replaceAll("'", "&apos;");
}

export function renderCandlestickSVG(
  symbol: string,
  timeframe: string,
  candles: Candle[],
  width = 1240,
  height = 620
): string | null {
  if (candles.length === 0) {
    return null;
  }

  const sorted = [...candles].sort((a, b) => Date.parse(a.bucketStart) - Date.parse(b.bucketStart));
  const safeSymbol = escapeXml(symbol.toUpperCase());
  const safeTimeframe = escapeXml(timeframe.toLowerCase());

  const bg = "#0B0E11";
  const fg = "#EAECEF";
  const muted = "#7B8596";
  const grid = "#1E2329";
  const up = "#0ECB81";
  const down = "#F6465D";

  const chartLeft = 52;
  const chartRight = width - 24;
  const chartTop = 96;
  const chartBottom = height - 30;
  const chartWidth = Math.max(120, chartRight - chartLeft);
  const chartHeight = Math.max(120, chartBottom - chartTop);

  const minLow = Math.min(...sorted.map((candle) => candle.low));
  const maxHigh = Math.max(...sorted.map((candle) => candle.high));
  const span = Math.max(0.0001, maxHigh - minLow);

  const priceToY = (value: number): number => chartBottom - ((value - minLow) / span) * chartHeight;

  const step = chartWidth / sorted.length;
  const bodyWidth = Math.max(2, Math.min(9, step * 0.66));

  const gridLines = Array.from({ length: 6 }, (_, idx) => {
    const y = chartTop + (chartHeight / 5) * idx;
    return `<line x1='${chartLeft}' y1='${y.toFixed(2)}' x2='${chartRight}' y2='${y.toFixed(2)}' stroke='${grid}' stroke-width='1' opacity='0.6'/>`;
  }).join("");

  const candlesSvg = sorted
    .map((candle, idx) => {
      const x = chartLeft + (idx + 0.5) * step;
      const yHigh = priceToY(candle.high);
      const yLow = priceToY(candle.low);
      const yOpen = priceToY(candle.open);
      const yClose = priceToY(candle.close);
      const color = candle.close >= candle.open ? up : down;
      const bodyTop = Math.min(yOpen, yClose);
      const bodyHeight = Math.max(1.8, Math.abs(yClose - yOpen));
      const bodyX = x - bodyWidth / 2;

      return [
        `<line data-candle-idx='${idx}' class='candle-wick' x1='${x.toFixed(2)}' y1='${yHigh.toFixed(2)}' x2='${x.toFixed(2)}' y2='${yLow.toFixed(2)}' stroke='${color}' stroke-width='1.2'/>`,
        `<rect data-candle-idx='${idx}' class='candle-body' x='${bodyX.toFixed(2)}' y='${bodyTop.toFixed(2)}' width='${bodyWidth.toFixed(2)}' height='${bodyHeight.toFixed(2)}' fill='${color}' rx='1'/>`
      ].join("");
    })
    .join("");

  return [
    `<svg xmlns='http://www.w3.org/2000/svg' width='${width}' height='${height}' viewBox='0 0 ${width} ${height}'>`,
    `<rect width='${width}' height='${height}' fill='${bg}'/>`,
    `<text x='24' y='44' fill='${fg}' font-size='24' font-family='monospace'>${safeSymbol}</text>`,
    `<text x='24' y='74' fill='${muted}' font-size='16' font-family='monospace'>timeframe ${safeTimeframe}</text>`,
    `<rect x='${chartLeft}' y='${chartTop}' width='${chartWidth}' height='${chartHeight}' fill='none' stroke='${grid}' stroke-width='1'/>`,
    gridLines,
    candlesSvg,
    "</svg>"
  ].join("");
}
