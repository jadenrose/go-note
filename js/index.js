// global htmx

const POST = 'post';
const PUT = 'put';

htmx.on('htmx:beforeSwap', (e) => {
	if (e.detail.xhr.status === 422) {
		switch (e.detail.requestConfig.verb) {
			// On semantic error, replace with original
			case PUT: {
				e.detail.shouldSwap = true;
				e.detail.isError = false;
				break;
			}
			// On semantic error, delete the temp field
			case POST: {
				e.detail.shouldSwap = true;
				e.detail.isError = false;
				e.detail.swapOverride = 'delete';
				break;
			}
		}
	}
});

htmx.on('htmx:load', (e) => {
	if (e.detail.elt.id === 'search-results-list') {
		const searchTerm =
			e.detail.elt.attributes.getNamedItem('data-search-term')?.value;

		if (!searchTerm) {
			return;
		}

		const words = searchTerm.split(/\s/).filter((s) => !!s);
		const regex = new RegExp(words.join('|'), 'gi');
		const elements = document.querySelectorAll(
			'.search-result-title, .search-result-content',
		);

		for (const el of elements) {
			el.innerHTML = el.textContent.replace(
				regex,
				(match) => `<span class="highlight">${match}</span>`,
			);
		}
	}
});

window.onload = () => {
	const input = document.getElementById('search-term');
	document.addEventListener('keydown', (e) => {
		if (e.key === 'F' && e.ctrlKey && e.shiftKey) {
			input.focus();
			input.select();
			return;
		}

		if (document.activeElement === input && e.key === 'Escape') {
			input.blur();
			return;
		}
	});
};
