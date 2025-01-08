document.addEventListener('DOMContentLoaded', init, false);

const tmpl = {
    results: Handlebars.compile(`
        <table class="table table-hover">
          <thead class="thead-dark">
            <tr>
              <th class="sorting text-nowrap" scope="col"><a onclick="updateSorting('reference')">Aktenzeichen {{arrow 'reference'}}</a></th>
              <th class="sorting text-nowrap" scope="col"><a onclick="updateSorting('documentType')">Form {{arrow 'documentType'}}</a></th>
              <th class="sorting text-nowrap" scope="col"><a onclick="updateSorting('date')">Datum {{arrow 'date'}}</a></th>
              <th class="sorting text-nowrap" scope="col"><a onclick="updateSorting('decision')">Entscheidung {{arrow 'decision'}}</a></th>
              <th scope="col">BE/ERi</th>
              <th scope="col">Sachgebiet</th>
              <th scope="col">Gegenstand</th>
              <th scope="col">Schlagworte</th>
              <th scope="col">Kommentare</th>
            </tr>
          </thead>
          <tbody id="results">
            {{#each this}}
            <tr>
                <th class="text-nowrap" scope="row"><a href="/download/{{this.id}}">{{this.reference}}</a></th>
                <td class="text-nowrap">{{this.documentType}}</td>
                <td class="text-nowrap">{{this.date}}</td>
                <td class="text-nowrap">{{this.decision}}</td>
                <td class="text-nowrap">{{this.authorType}}:&nbsp;{{this.author}}</td>
                <td class="text-nowrap">{{{this.area}}}</td>
                <td>{{this.subject}}</td>
                <td>
                {{#each this.keywords}}
                {{this}}{{#unless @last}}<br>{{/unless}}
                {{/each}}
                </td>
                <td>
                {{#each this.comments}}
                {{this}}{{#unless @last}}<br>{{/unless}}
                {{/each}}
                </td>
            </tr>
            {{/each}}
            <tr><td></td><td></td><td></td><td></td><td></td><td></td><td></td><td></td><td></td></tr>
          </tbody>
        </table>
    `),
    autocomplete: Handlebars.compile(`
        {{#each this}}
            <div class="dropdown-item {{this.active}}" onclick="addTag({{this.index}});">
                <span class="text-secondary">{{this.key}}</span>
                <span>{{{this.prefix}}}<b>{{{this.prompt}}}</b>{{{this.suffix}}}</span>
            </div>
        {{/each}}
    `),
    tags: Handlebars.compile(`
        {{#each this}}
            <span class="badge badge-secondary font-weight-normal">
                <span>{{this.key}}&nbsp;<b>{{{this.value}}}</b>&nbsp;</span><span class="x" onclick="removeTag({{this.index}});">&#x2715;</span>
            </span>
        {{/each}}
    `),
};

var state = {
    sorting: {
        column: 'date',
        asc: false,
    },
    tags: [],
    autocomplete: [],
    queryTags: new Set(),
    results: null,
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
    ui.input.addEventListener('blur', evt => {
        if (evt.relatedTarget != ui.autocomplete) {
            ui.autocomplete.classList.remove('show');
        }
    });
    ui.input.addEventListener('keypress', evt => {
        if (evt.key == 'Enter') {
            evt.preventDefault();
            search();
        }
    });
    ui.autocomplete.addEventListener('blur', evt => {
        ui.autocomplete.classList.remove('show');
    });

    Handlebars.registerHelper('arrow', col => {
        if (state.sorting.column != col) {
            return '';
        }
        return state.sorting.asc ? '▲' : '▼';
    });

    search();
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
    fn = (a, b) => (state.sorting.asc ? 1 : -1) * a[state.sorting.column].localeCompare(b[state.sorting.column]);
    if (state.sorting.column == 'date') {
        fn = (a, b) => {
            [dayA, monthA, yearA] = a.date.split('.');
            [dayB, monthB, yearB] = b.date.split('.');
            return (state.sorting.asc ? 1 : -1) * (new Date(yearA, monthA-1, dayA) - new Date(yearB, monthB-1, dayB));
        }
    }
    state.results = state.results.sort(fn);
    ui.results.innerHTML = tmpl.results(state.results);
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
        state.results = data;
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

function updateSorting(col) {
    if (state.sorting.column == col) {
        state.sorting.asc = !state.sorting.asc;
    } else {
        state.sorting.column = col;
        state.sorting.asc = true;
    }
    renderResults();
}