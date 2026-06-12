// Shared arrow-drawing utility for Level 3 diagrams

function initSvg(svg, ref) {
  svg.innerHTML = '<defs></defs>';
  svg.setAttribute('width',  ref.offsetWidth);
  svg.setAttribute('height', ref.offsetHeight);
}

function getEdge(el, ref, side) {
  const e = el.getBoundingClientRect();
  const r = ref.getBoundingClientRect();
  const x = e.left - r.left;
  const y = e.top  - r.top;
  const w = e.width, h = e.height;
  switch (side) {
    case 'top':    return { x: x + w/2, y: y };
    case 'bottom': return { x: x + w/2, y: y + h };
    case 'left':   return { x: x,       y: y + h/2 };
    case 'right':  return { x: x + w,   y: y + h/2 };
  }
}

function ensureMarker(svg, color) {
  const id = 'mk' + color.replace(/[^a-z0-9]/gi, '');
  if (!svg.querySelector('#' + id)) {
    const defs = svg.querySelector('defs');
    defs.innerHTML += `<marker id="${id}" markerWidth="7" markerHeight="7"
      refX="5" refY="3" orient="auto">
      <path d="M0,0 L0,6 L7,3 z" fill="${color}" opacity="0.85"/>
    </marker>`;
  }
  return id;
}

function arrow(svg, ref, fromId, fromSide, toId, toSide, color, dashed, label) {
  const fromEl = document.getElementById(fromId);
  const toEl   = document.getElementById(toId);
  if (!fromEl || !toEl) return;

  const p1 = getEdge(fromEl, ref, fromSide);
  const p2 = getEdge(toEl,   ref, toSide);
  const mid = { x: (p1.x + p2.x) / 2, y: (p1.y + p2.y) / 2 };

  const mkId = ensureMarker(svg, color);

  const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
  line.setAttribute('x1', p1.x); line.setAttribute('y1', p1.y);
  line.setAttribute('x2', p2.x); line.setAttribute('y2', p2.y);
  line.setAttribute('stroke', color);
  line.setAttribute('stroke-width', '1.5');
  if (dashed) line.setAttribute('stroke-dasharray', '6,4');
  line.setAttribute('marker-end', `url(#${mkId})`);
  svg.appendChild(line);

  if (label) {
    const tw = label.length * 5.5 + 10;
    const bg = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
    bg.setAttribute('x', mid.x - tw/2); bg.setAttribute('y', mid.y - 8);
    bg.setAttribute('width', tw); bg.setAttribute('height', 14);
    bg.setAttribute('fill', '#0f1117'); bg.setAttribute('rx', 3);
    svg.appendChild(bg);

    const txt = document.createElementNS('http://www.w3.org/2000/svg', 'text');
    txt.setAttribute('x', mid.x); txt.setAttribute('y', mid.y + 2);
    txt.setAttribute('text-anchor', 'middle');
    txt.setAttribute('fill', color);
    txt.setAttribute('font-size', '10');
    txt.setAttribute('font-family', 'Segoe UI, system-ui, sans-serif');
    txt.setAttribute('opacity', '0.85');
    txt.textContent = label;
    svg.appendChild(txt);
  }
}
