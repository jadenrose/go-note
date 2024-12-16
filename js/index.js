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
