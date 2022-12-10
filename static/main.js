document.addEventListener('DOMContentLoaded', init, false);

const tmpl = {
    results: Handlebars.compile(`
        {{#each this}}
        <tr>
            <th class="text-nowrap" scope="row"><a href="/download/{{this.document.id}}">{{this.document.reference}}</a></th>
            <td class="text-nowrap">{{this.document.documentType}}</td>
            <td class="text-nowrap">{{this.document.date}}</td>
            <td class="text-nowrap">{{this.document.decision}}</td>
            <td class="text-nowrap">{{this.document.authorType}}:&nbsp;{{this.document.author}}</td>
            <td class="text-nowrap">{{this.document.area}}</td>
            <td>{{this.document.subject}}</td>
            <td>
            {{#each this.document.keywords}}
            {{this}}<br>
            {{/each}}
            </td>
            <td>
            {{#each this.document.comments}}
            {{this}}<br>
            {{/each}}
            </td>
        </tr>
        {{/each}}
        <tr><td></td><td></td><td></td><td></td><td></td><td></td><td></td><td></td><td></td></tr>
    `),
    autocomplete: Handlebars.compile(`
        {{#each this}}
            <div class="dropdown-item {{this.active}}" onclick="addTag({{this.index}});">
                <span class="text-secondary">{{this.key}}</span>
                <span>{{this.prefix}}<b>{{this.prompt}}</b>{{this.suffix}}</span>
            </div>
        {{/each}}
    `),
    tags: Handlebars.compile(`
        {{#each this}}
            <span class="badge badge-secondary font-weight-normal">
                <span>{{this.key}}&nbsp;<b>{{this.value}}</b>&nbsp;</span><span class="x" onclick="removeTag({{this.index}});">&#x2715;</span>
            </span>
        {{/each}}
    `),
};

var state = {
    tags: [],
    autocomplete: [],
    queryTags: new Set(),
    result: null,
};
var ui;

function init() {
    fetch('tags', {
        method: 'GET',
        cache: 'no-cache',
    }).then((response) => response.json()).then((data) => state.tags = data);

    ui = {
        input: document.getElementById('search-box'),
        autocomplete: document.getElementById('autocomplete'),
        results: document.getElementById('results'),
        tags: document.getElementById('selected-tags'),
    };

    ui.input.addEventListener('input', () => {
        updateAutocomplete();
        renderAutocomplete();
        search();
    });
    ui.input.addEventListener('mouseup', () => {
        updateAutocomplete();
        renderAutocomplete();
    });
    ui.input.addEventListener('blur', (evt) => {
        if (evt.relatedTarget != ui.autocomplete) {
            ui.autocomplete.classList.remove('show');
        }
    });
    ui.input.addEventListener('keypress', (evt) => {
        if (evt.key == 'Enter') {
            evt.preventDefault();
            search();
        }
    });

    ui.autocomplete.addEventListener('blur', (evt) => {
        ui.autocomplete.classList.remove('show');
    });
}

function renderAutocomplete() {
    if (state.autocomplete.length == 0) {
        ui.autocomplete.classList.remove('show');
        return;
    }
    ui.autocomplete.innerHTML = tmpl.autocomplete(state.autocomplete);
    ui.autocomplete.classList.add('show');
}

function renderTags() {
    tags = [];
    for (const idx of state.queryTags) {
        tag = state.tags[idx];
        tags.push({
            index: idx,
            key: tag.key,
            value: tag.value,
        });
    }
    ui.tags.innerHTML = tmpl.tags(tags);
}

function renderResults() {
    ui.results.innerHTML = tmpl.results(state.result);
}

function search() {
    tags = [];
    for (const idx of state.queryTags) {
        tags.push(state.tags[idx]);
    }
    fetch('search', {
        method: 'POST',
        cache: 'no-cache',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            query: ui.input.value,
            tags: tags,
        }),
    }).then((response) => response.json()).then((data) => {
        state.result = data;
        renderResults();
    });
}

function updateAutocomplete() {
    state.autocomplete = [];

    if (ui.input.selectionStart != ui.input.selectionEnd) {
        return;
    }

    query = ui.input.value;
    caretPos = ui.input.selectionStart;
    word = (query.slice(0, caretPos).match(/\S*$/) + query.slice(caretPos).match(/^\S*/)).toLowerCase();

    if (word == '') {
        return;
    }

    for ([idx, tag] of state.tags.entries()) {
        pos = tag.value.toLowerCase().indexOf(word);
        if (pos != -1) {
            state.autocomplete.push({
                'index': idx,
                'pos': pos,
                'prefix': tag.value.slice(0, pos),
                'prompt': tag.value.slice(pos, pos + word.length),
                'suffix': tag.value.slice(pos + word.length),
                'key': tag.key,
            })
        }
    }
    state.autocomplete.sort((a, b) => a.pos - b.pos);
}

function addTag(idx) {
    state.queryTags.add(idx);
    renderTags(); 

    query = ui.input.value;
    caretPos = ui.input.selectionStart;
    left = query.slice(0, caretPos).match(/\S*$/);
    right = query.slice(caretPos).match(/^\S*/);
    ui.input.value = query.slice(0, left.index) + query.slice(caretPos + right[0].length);
    ui.input.selectionStart = caretPos - left[0].length;
    ui.input.selectionEnd = caretPos - left[0].length;
    ui.input.focus();
    ui.autocomplete.classList.remove('show');
    search();
}

function removeTag(idx) {
    state.queryTags.delete(idx);
    renderTags();
    search();
}
