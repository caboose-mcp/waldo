// Initialize CodeMirror editor
const editor = CodeMirror.fromTextArea(document.getElementById('meml-editor'), {
  mode: 'properties',
  theme: 'monokai',
  lineNumbers: true,
  indentUnit: 2,
  tabSize: 2,
  indentWithTabs: false,
  lineWrapping: true,
  styleActiveLine: true,
});

// Tone slider elements
const sliders = {
  formality: { slider: document.getElementById('formality'), value: document.getElementById('formality-value') },
  directness: { slider: document.getElementById('directness'), value: document.getElementById('directness-value') },
  humor: { slider: document.getElementById('humor'), value: document.getElementById('humor-value') },
  hedging: { slider: document.getElementById('hedging'), value: document.getElementById('hedging-value') },
  warmth: { slider: document.getElementById('warmth'), value: document.getElementById('warmth-value') },
};

// Example personas
const examples = {
  default: {
    meml: `# default persona

[🪪 meta]
name        = "default"
version     = "0.1.0"

[🎭 tone]
formality   = 0.5
directness  = 0.5
humor       = 0.5
hedging     = 0.5
warmth      = 0.5`,
    tone: { formality: 0.5, directness: 0.5, humor: 0.5, hedging: 0.5, warmth: 0.5 }
  },
  direct: {
    meml: `# direct persona

[🪪 meta]
name        = "direct"
version     = "0.1.0"

[🎭 tone]
formality   = 0.3
directness  = 0.9
humor       = 0.3
hedging     = 0.1
warmth      = 0.4`,
    tone: { formality: 0.3, directness: 0.9, humor: 0.3, hedging: 0.1, warmth: 0.4 }
  },
  formal: {
    meml: `# formal persona

[🪪 meta]
name        = "formal"
version     = "0.1.0"

[🎭 tone]
formality   = 0.9
directness  = 0.6
humor       = 0.2
hedging     = 0.7
warmth      = 0.3`,
    tone: { formality: 0.9, directness: 0.6, humor: 0.2, hedging: 0.7, warmth: 0.3 }
  },
  warm: {
    meml: `# warm persona

[🪪 meta]
name        = "warm"
version     = "0.1.0"

[🎭 tone]
formality   = 0.4
directness  = 0.5
humor       = 0.7
hedging     = 0.4
warmth      = 0.9`,
    tone: { formality: 0.4, directness: 0.5, humor: 0.7, hedging: 0.4, warmth: 0.9 }
  },
  technical: {
    meml: `# technical persona

[🪪 meta]
name        = "technical"
version     = "0.1.0"

[🎭 tone]
formality   = 0.7
directness  = 0.8
humor       = 0.2
hedging     = 0.5
warmth      = 0.2`,
    tone: { formality: 0.7, directness: 0.8, humor: 0.2, hedging: 0.5, warmth: 0.2 }
  }
};

// Simple MEML parser
function parseMeml(memlText) {
  const result = {
    meta: { name: 'persona', version: '0.1.0' },
    tone: {}
  };

  let section = null;
  const lines = memlText.split('\n');

  for (const line of lines) {
    const trimmed = line.trim();

    // Skip comments and empty lines
    if (!trimmed || trimmed.startsWith('#')) continue;

    // Section headers like [🪪 meta] or [🎭 tone]
    if (trimmed.startsWith('[') && trimmed.endsWith(']')) {
      // Extract section name (remove emoji and spaces)
      section = trimmed.replace(/[\[\]🪪🎭📢🗣️\s]/g, '').toLowerCase();
      continue;
    }

    // Parse key = value
    const eqIndex = trimmed.indexOf('=');
    if (eqIndex === -1) continue;

    const key = trimmed.substring(0, eqIndex).trim();
    let value = trimmed.substring(eqIndex + 1).trim();

    // Remove quotes
    if ((value.startsWith('"') && value.endsWith('"')) ||
        (value.startsWith("'") && value.endsWith("'"))) {
      value = value.slice(1, -1);
    }

    // Try to parse as number
    if (!isNaN(value) && value !== '') {
      value = parseFloat(value);
    }

    // Store in appropriate section
    if (section === 'meta') {
      result.meta[key] = value;
    } else if (section === 'tone') {
      result.tone[key] = value;
    }
  }

  return result;
}

// Generate ChatGPT system prompt
function generateChatGPTPrompt(parsed) {
  const { tone } = parsed;

  let formality = '';
  if (tone.formality > 0.7) {
    formality = 'Use formal, professional language. Avoid slang and contractions.';
  } else if (tone.formality > 0.4) {
    formality = 'Use moderately formal language.';
  } else {
    formality = 'Use casual, conversational language. Contractions and informal expressions are fine.';
  }

  let directness = '';
  if (tone.directness > 0.7) {
    directness = 'Be direct and concise. Get to the point quickly without unnecessary elaboration.';
  } else if (tone.directness > 0.4) {
    directness = 'Be reasonably direct while providing adequate context.';
  } else {
    directness = 'Take time to provide context and perspective. Lead gently toward conclusions.';
  }

  let humor = '';
  if (tone.humor > 0.7) {
    humor = 'Use humor, wit, and wordplay frequently throughout responses.';
  } else if (tone.humor > 0.3) {
    humor = 'Use light humor and occasional wit where appropriate.';
  } else {
    humor = 'Keep humor minimal. Maintain a serious, straightforward tone.';
  }

  let hedging = '';
  if (tone.hedging > 0.7) {
    hedging = 'Use qualifying language extensively. Include caveats, "may," "might," and uncertainty markers.';
  } else if (tone.hedging > 0.3) {
    hedging = 'Use some qualifying language when appropriate, but be reasonably confident.';
  } else {
    hedging = 'Be confident in your responses. Avoid over-qualifying unless there is genuine uncertainty.';
  }

  let warmth = '';
  if (tone.warmth > 0.7) {
    warmth = 'Be warm, encouraging, and enthusiastic. Show genuine interest in helping.';
  } else if (tone.warmth > 0.3) {
    warmth = 'Be friendly and approachable while maintaining professionalism.';
  } else {
    warmth = 'Maintain a neutral, professional demeanor. Focus on facts over emotional connection.';
  }

  return `You are a helpful assistant with the following personality and communication style:

## Formality
${formality}

## Directness
${directness}

## Humor
${humor}

## Confidence
${hedging}

## Warmth
${warmth}

Apply these guidelines consistently across all your responses. Adapt them to the context when needed, but maintain this overall voice.`;
}

// Generate Gemini system prompt
function generateGeminiPrompt(parsed) {
  const { tone } = parsed;

  const traits = [];

  if (tone.formality > 0.7) traits.push('formal and professional');
  else if (tone.formality > 0.4) traits.push('moderately formal');
  else traits.push('casual and conversational');

  if (tone.directness > 0.7) traits.push('direct and concise');
  else if (tone.directness > 0.4) traits.push('reasonably direct');
  else traits.push('contextual and roundabout');

  if (tone.humor > 0.7) traits.push('witty and humorous');
  else if (tone.humor > 0.3) traits.push('lightly humorous');
  else traits.push('serious and dry');

  if (tone.hedging > 0.7) traits.push('cautious with many qualifiers');
  else if (tone.hedging > 0.3) traits.push('appropriately confident');
  else traits.push('confident and assertive');

  if (tone.warmth > 0.7) traits.push('warm and encouraging');
  else if (tone.warmth > 0.3) traits.push('friendly');
  else traits.push('neutral and detached');

  return `You are a helpful AI assistant. Respond in a voice that is: ${traits.join(', ')}.

Your tone values:
- Formality: ${(tone.formality || 0.5).toFixed(2)}
- Directness: ${(tone.directness || 0.5).toFixed(2)}
- Humor: ${(tone.humor || 0.5).toFixed(2)}
- Hedging: ${(tone.hedging || 0.5).toFixed(2)}
- Warmth: ${(tone.warmth || 0.5).toFixed(2)}`;
}

// Convert parsed object to JSON
function toJSON(parsed) {
  return JSON.stringify(parsed, null, 2);
}

// Update all outputs
function updateOutputs() {
  const memlText = editor.getValue();

  try {
    const parsed = parseMeml(memlText);

    // Update JSON output
    document.getElementById('json-output').textContent = toJSON(parsed);

    // Update ChatGPT prompt
    document.getElementById('prompt-output').textContent = generateChatGPTPrompt(parsed);

    // Update Gemini prompt
    document.getElementById('gemini-output').textContent = generateGeminiPrompt(parsed);

    // Update validation
    document.getElementById('validation-output').textContent = '✓ Valid MEML config';

    // Update sliders based on parsed tone
    for (const [key, value] of Object.entries(parsed.tone || {})) {
      if (sliders[key]) {
        const percent = Math.round(value * 100);
        sliders[key].slider.value = percent;
        sliders[key].value.textContent = value.toFixed(2);
      }
    }
  } catch (err) {
    document.getElementById('validation-output').textContent = `❌ Parse error:\n${err.message}`;
  }
}

// Generate MEML from sliders
function generateMemlFromSliders() {
  const tone = {
    formality: parseInt(sliders.formality.slider.value) / 100,
    directness: parseInt(sliders.directness.slider.value) / 100,
    humor: parseInt(sliders.humor.slider.value) / 100,
    hedging: parseInt(sliders.hedging.slider.value) / 100,
    warmth: parseInt(sliders.warmth.slider.value) / 100,
  };

  const parsed = parseMeml(editor.getValue());
  const name = parsed.meta.name || 'persona';

  let meml = `# ${name}\n\n`;
  meml += `[🪪 meta]\n`;
  meml += `name        = "${name}"\n`;
  meml += `version     = "0.1.0"\n\n`;
  meml += `[🎭 tone]\n`;
  for (const [key, value] of Object.entries(tone)) {
    meml += `${key} = ${value.toFixed(2)}\n`;
  }

  return meml;
}

// Update tone value displays and MEML editor
function updateSlider(key) {
  const percent = parseInt(sliders[key].slider.value);
  const value = (percent / 100).toFixed(2);
  sliders[key].value.textContent = value;

  // Update MEML in editor
  const current = editor.getValue();
  const updated = current.replace(
    new RegExp(`^${key}\\s*=\\s*[\\d.]+`, 'm'),
    `${key}   = ${value}`
  );
  editor.setValue(updated);
  updateOutputs();
}

// Slider event listeners
for (const [key, elements] of Object.entries(sliders)) {
  elements.slider.addEventListener('input', () => updateSlider(key));
}

// Example buttons
document.querySelectorAll('[data-example]').forEach(btn => {
  btn.addEventListener('click', () => {
    const example = btn.dataset.example;
    if (examples[example]) {
      editor.setValue(examples[example].meml);
      updateOutputs();
    }
  });
});

// Tab switching
document.querySelectorAll('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    const tabName = tab.dataset.tab;

    // Update active tab button
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    tab.classList.add('active');

    // Update active content
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    document.getElementById(`${tabName}-content`).classList.add('active');
  });
});

// Copy buttons
document.getElementById('copy-json').addEventListener('click', () => {
  const json = document.getElementById('json-output').textContent;
  navigator.clipboard.writeText(json).then(() => {
    showMessage('Copied JSON to clipboard', 'success');
  }).catch(() => {
    showMessage('Failed to copy. Try right-click → Copy', 'error');
  });
});

document.getElementById('copy-prompt').addEventListener('click', () => {
  const prompt = document.getElementById('prompt-output').textContent;
  navigator.clipboard.writeText(prompt).then(() => {
    showMessage('Copied ChatGPT prompt to clipboard', 'success');
  }).catch(() => {
    showMessage('Failed to copy. Try right-click → Copy', 'error');
  });
});

document.getElementById('copy-gemini').addEventListener('click', () => {
  const prompt = document.getElementById('gemini-output').textContent;
  navigator.clipboard.writeText(prompt).then(() => {
    showMessage('Copied Gemini prompt to clipboard', 'success');
  }).catch(() => {
    showMessage('Failed to copy. Try right-click → Copy', 'error');
  });
});

document.getElementById('download-meml').addEventListener('click', () => {
  const meml = generateMemlFromSliders();
  const parsed = parseMeml(editor.getValue());
  const filename = `${parsed.meta.name || 'persona'}.meml`;

  const blob = new Blob([meml], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);

  showMessage(`Downloaded ${filename}`, 'success');
});

// Show message
function showMessage(text, type = 'info') {
  const msg = document.getElementById('message');
  msg.textContent = text;
  msg.className = `message show ${type}`;
  setTimeout(() => msg.classList.remove('show'), 3000);
}

// Add Gemini output content area (hidden in HTML but referenced in JS)
if (!document.getElementById('gemini-output')) {
  const tabContent = document.createElement('div');
  tabContent.className = 'tab-content';
  tabContent.id = 'gemini-content';
  const pre = document.createElement('pre');
  pre.id = 'gemini-output';
  tabContent.appendChild(pre);
  document.querySelector('.preview-pane').insertBefore(tabContent, document.querySelector('.actions'));

  // Add Gemini tab button
  const geminiTab = document.createElement('button');
  geminiTab.className = 'tab';
  geminiTab.dataset.tab = 'gemini';
  geminiTab.textContent = 'Gemini Prompt';
  geminiTab.addEventListener('click', () => {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    geminiTab.classList.add('active');
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    tabContent.classList.add('active');
  });
  document.querySelector('.tabs').appendChild(geminiTab);
}

// Initial render
updateOutputs();

// Update on editor changes (with debounce)
let updateTimeout;
editor.on('change', () => {
  clearTimeout(updateTimeout);
  updateTimeout = setTimeout(updateOutputs, 300);
});
