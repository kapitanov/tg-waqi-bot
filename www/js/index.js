const ResultRowPresenter = {
    template: `
<li v-if="!!value">
    <strong>{{ name }}: </strong>
    <span>{{ value }}</span>
</li>
`,
    props: ['name', 'value']
}
Vue.component('v-result-row', ResultRowPresenter);

const ResultPresenter = {
    template: `
<div>
    <div :class="getCardCssClass()" v-if="!!result">
        <div class="card-header">
            <h5 class="card-title"></h5>Air Quality: {{ getTitle() }}</h5>
        </div>
        <ul class="list-unstyled m-3">
            <li>
                <strong>Station:</strong>
                #{{ result?.station?.id }} &quot;{{ result?.station?.name }}&quot;
            </li>
            <v-result-row name="Air quality index" v-bind:value="result?.aqi"></v-result-row>
            <v-result-row name="Particulate matter 2.5" v-bind:value="result?.pm25"></v-result-row>
            <v-result-row name="Particulate matter 10" v-bind:value="result?.pm10"></v-result-row>
            <v-result-row name="Ozone" v-bind:value="result?.o3"></v-result-row>
            <v-result-row name="Nitrogen dioxide" v-bind:value="result?.no2"></v-result-row>
            <v-result-row name="Sulfur dioxide" v-bind:value="result?.so2"></v-result-row>
            <v-result-row name="Carbon monoxide" v-bind:value="result?.co"></v-result-row>
            <v-result-row name="Last updated" v-bind:value="result?.time"></v-result-row>
        </ul>
    </div>

    <div class="mt-4" v-if="!!loading">
        <div class="spinner-border" role="status">
            <span class="visually-hidden">Loading...</span>
        </div>
    </div>

    <div class="alert alert-danger alert-dismissible mt-4" role="alert" v-if="!!error">
        <strong>Error</strong> {{ error }}
        <button type="button" class="btn-close" v-on:click="dismiss"></button>
    </div>
</div>
`,
    props: {
        error: String,
        loading: Boolean,
        result: Object,
    },

    methods: {
        dismiss() {
            this.$parent.clearError();
        },

        getCardCssClass() {
            return `card mt-4 border-${this.getCss()}`;
        },

        getCss() {
            switch (this.result?.level) {
                case 'good':
                    return 'success';
                case 'moderate':
                    return 'warning';
                case 'possibly_unhealthy':
                    return 'warning';
                case 'unhealthy':
                    return 'danger';
                case 'very_unhealthy':
                    return 'danger';
                case 'hazardous':
                    return 'danger';
                default:
                    return '';
            }
        },

        getTitle() {
            switch (this.result?.level) {
                case 'good':
                    return 'Good';
                case 'moderate':
                    return 'Moderate';
                case 'possibly_unhealthy':
                    return 'Possibly unhealthy';
                case 'unhealthy':
                    return 'Unhealthy';
                case 'very_unhealthy':
                    return 'Very unhealthy';
                case 'hazardous':
                    return 'Hazardous';
                default:
                    return '';
            }
        }
    }
}
Vue.component('v-result-presenter', ResultPresenter);

const ViewByLocation = {
    template: `
<div>
    <form v-on:submit.prevent="submit" class="row mt-4">
        <div class="col-sm-12">
            <label class="form-label">Geo coordinates (latitude/longitude):</label>
        </div>
        <div class="col-sm-6">
            <input type="number" class="form-control" min="0" value="0" step="0.000001" v-model="lat" :disabled="loading" placeholder="Latitude">
        </div>
        <div class="col-sm-6">
            <input type="number" class="form-control" min="0" value="0" step="0.000001" v-model="lon" :disabled="loading" placeholder="Longitude">
        </div>
        <div class="col-sm-12 mt-4">
            <button type="submit" class="btn btn-primary" :disabled="loading">
                <i class="bi bi-caret-right-fill"></i> Go
            </button>
            <button type="button" class="btn btn-secondary" v-on:click="geo" :disabled="loading">
                <i class="bi bi-geo-alt-fill"></i> Use current location
            </button>
        </div>
    </form>

    <v-result-presenter v-bind:error="error" v-bind:loading="loading" v-bind:result="result" />
</div>
`,
    data() {
        return {
            lon: null,
            lat: null,
            error: null,
            loading: false,
            result: null,
        }
    },

    methods: {
        submit() {
            if (!this.lat || !this.lon) {
                this.error = 'Missing coordinates';
                return;
            }

            this.error = null;
            this.result = null;
            this.loading = true;
            const self = this;

            fetch(`/api/status/geo?lat=${this.lat}&lon=${this.lon}`)
                .then((response) => {
                    return response.json();
                })
                .then((result) => {
                    self.loading = false;
                    self.result = result;
                })
                .catch((e) => {
                    self.loading = false;
                    self.error = e;
                    self.state = null;
                });
        },

        geo() {
            const self = this;
            navigator.geolocation.getCurrentPosition((position) => {
                self.lat = Math.round(position.coords.latitude * 100000.0) / 100000.0;
                self.lon = Math.round(position.coords.longitude * 100000.0) / 100000.0;
            })
        },

        clearError() {
            this.error = null;
        }
    }
};
Vue.component('v-view-by-location', ViewByLocation);


const ViewByCity = {
    template: `
<div>
    <form v-on:submit.prevent="submit" class="row mt-4">
        <div class="col-sm-12">
            <label class="form-label">City:</label>
        </div>
        <div class="col-sm-12">
            <input type="text" class="form-control" v-model="city" :disabled="loading" placeholder="City">
        </div>
        <div class="col-sm-12 mt-4">
            <button type="submit" class="btn btn-primary" :disabled="loading">
                <i class="bi bi-caret-right-fill"></i> Go
            </button>
        </div>
    </form>

    <v-result-presenter v-bind:error="error" v-bind:loading="loading" v-bind:result="result" />
</div>
`,
    data() {
        return {
            city: null,
            error: null,
            loading: false,
            result: null,
        }
    },

    methods: {
        submit() {
            if (!this.city) {
                this.error = 'Missing city';
                return;
            }

            this.error = null;
            this.result = null;
            this.loading = true;
            const self = this;

            fetch(`/api/status/city/${encodeURIComponent(this.city)}`)
                .then((response) => {
                    return response.json();
                })
                .then((result) => {
                    self.loading = false;
                    self.result = result;
                })
                .catch((e) => {
                    self.loading = false;
                    self.error = e;
                    self.state = null;
                });
        },

        clearError() {
            this.error = null;
        }
    }
};
Vue.component('v-view-by-city', ViewByCity);


const ViewByStation = {
    template: `
<div>
    <form v-on:submit.prevent="submit" class="row mt-4">
        <div class="col-sm-12">
            <label class="form-label">Station ID:</label>
        </div>
        <div class="col-sm-12">
            <input type="number" class="form-control" v-model="station" :disabled="loading" placeholder="Station ID">
        </div>
        <div class="col-sm-12 mt-4">
            <button type="submit" class="btn btn-primary" :disabled="loading">
                <i class="bi bi-caret-right-fill"></i> Go
            </button>
        </div>
    </form>

    <v-result-presenter v-bind:error="error" v-bind:loading="loading" v-bind:result="result" />
</div>
`,
    data() {
        return {
            station: null,
            error: null,
            loading: false,
            result: null,
        }
    },

    methods: {
        submit() {
            if (!this.station) {
                this.error = 'Missing station ID';
                return;
            }

            this.error = null;
            this.result = null;
            this.loading = true;
            const self = this;

            fetch(`/api/status/station/${this.station}`)
                .then((response) => {
                    return response.json();
                })
                .then((result) => {
                    self.loading = false;
                    self.result = result;
                })
                .catch((e) => {
                    self.loading = false;
                    self.error = e;
                    self.state = null;
                });
        },

        clearError() {
            this.error = null;
        }
    }
};
Vue.component('v-view-by-station', ViewByStation);


const App = {
    template: `
<div>
    <ul class="nav nav-tabs" role="tablist">
        <li class="nav-item">
            <button :class="getButtonClass('by-location')" type="button" v-on:click="selectPage('by-location')">
                By location
            </button>
        </li>
        <li class="nav-item">
            <button :class="getButtonClass('by-city')" type="button" v-on:click="selectPage('by-city')">
                By city
            </button>
        </li>
        <li class="nav-item">
            <button :class="getButtonClass('by-station')" type="button" v-on:click="selectPage('by-station')">
                By station
            </button>
        </li>
    </ul>
    <div class="tab-content" id="myTabContent">
        <div :class="getPaneClass('by-location')">
            <v-view-by-location></v-view-by-location>
        </div>
        <div :class="getPaneClass('by-city')">
            <v-view-by-city></v-view-by-city>
        </div>
        <div :class="getPaneClass('by-station')">
            <v-view-by-station></v-view-by-station>
        </div>
    </div>
</div>
`, data() {
        return {
            page: null,
        }
    },

    methods: {
        selectPage(page) {
            this.page = page;
        },

        getButtonClass(page) {
            return `nav-link${this.page === page ? ' active' : ''}`;
        },

        getPaneClass(page) {
            return `tab-pane fade${this.page === page ? ' show active' : ''}`;
        }
    },

    mounted() {
        this.selectPage('by-location');
    }
};
Vue.component('v-app', App);

window.addEventListener('load',
    () => {
        new Vue({ el: '#main' });
    });
