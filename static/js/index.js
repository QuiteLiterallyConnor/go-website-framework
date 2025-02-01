$(document).ready(function() {
    const audio = new Audio('https://storage.cloud.google.com/resume_connorisseur_com_music/Rainy_Music.mp3');
    audio.loop = true;
    audio.volume = 0.6;

    let jobs = [];
    let portfolioEntries = {};
    let contacts = {};
    let aboutContent = {};
    let recaptchaToken = '';

    function fetchJSON(file, callback) {
        return $.ajax({
            url: `/static/json/${file}`,
            method: 'GET',
            headers: {
                'X-Recaptcha-Token': recaptchaToken
            },
            success: callback,
            error: function() {
                console.error(`Error loading ${file}`);
            }
        });
    }

    function playAudio() {
        audio.play().catch(error => {
            console.error("Audio playback failed:", error);
        });
        document.removeEventListener('click', playAudio);
    }

    function toggleMute() {
        if (audio.muted) {
            audio.muted = false;
            $('#mute-icon').removeClass('bi-volume-mute').addClass('bi-volume-up');
            $('#mute-button').attr('title', 'Mute');
        } else {
            audio.muted = true;
            $('#mute-icon').removeClass('bi-volume-up').addClass('bi-volume-mute');
            $('#mute-button').attr('title', 'Unmute');
        }
    }

    function displayJobs() {
        const $resumeContent = $('#resume-content');
        $resumeContent.empty();

        jobs.forEach(job => {
            $resumeContent.append(`
                <br>
                <div class="job">
                    <p><strong>${job.position}</strong> <span style="float: right;"><strong>${job.duration}</strong></span></p>
                    <p>${job.company}</p>
                    <p>${job.responsibilities}</p>
                </div>
                <br>
            `);
        });
    }

    function displayPortfolio() {
        const $portfolioContent = $('#portfolio-content');
        $portfolioContent.empty();
        $portfolioContent.append(`<br><strong><a>A list of my most successful work in clients' videos. Credited as "Acaicia" in the videos' description</a></strong><br>`);

        ['3dAnimation', 'videoEditing'].forEach(section => {
            const sectionTitle = section === '3dAnimation' ? '3D Animation' : 'Video Editing';
            $portfolioContent.append(`<br><h2>${sectionTitle}</h2>`);

            portfolioEntries[section].forEach(entry => {
                $portfolioContent.append(`
                    <div class="portfolio-entry">
                        <p>
                            <strong><a href="${entry.videoLink}" target="_blank">${entry.videoName}</a></strong>
                            <span> | </span>
                            <a href="${entry.clientLink}" target="_blank">${entry.clientName}</a>
                            <span style="float: right;">${entry.viewCount}</span>
                        </p>
                        ${entry.gifs.map(gif => `<a href="${gif.link}" target="_blank"><img src="${gif.file}" alt="${entry.videoName}" class="portfolio-gif"></a>`).join('')}
                    </div>
                    <br>
                `);
            });
        });
    }

    function displayContacts() {
        const $content = $('#content');
        $content.html(`
            <h1>Contacts</h1>
            <div id="contacts-content">
                <br>
                <p><img src="/public/icons/email-icon.png" alt="Email" class="contact-icon"> ${contacts.email}</p>
                <p><img src="/public/icons/phone-icon.png" alt="Phone" class="contact-icon"> ${contacts.phone}</p>
            </div>
        `);
    }

    function displayAbout() {
        const $aboutContent = $('#box3');
        $aboutContent.empty();
        $aboutContent.append(`
            <h1>${aboutContent.title}</h1>
            <br>
            ${aboutContent.content.map(paragraph => `<p>${paragraph}</p>`).join('')}
        `);
        $aboutContent.append(`
            <br><br><br><br><br><br><br><br>
            <p>Hosted on THIS server</p>
            <img src="https://i.imgur.com/1BecPI3.jpeg" alt="Server" class="server-pic">
        `);
    }

    function initializeRecaptcha() {
        if (typeof grecaptcha !== 'undefined' && grecaptcha.enterprise) {
            grecaptcha.enterprise.ready(function() {
                grecaptcha.enterprise.execute('6LcTWBMqAAAAAEtY30zw0JD5hRFsjAu0ViwE3FiX', { action: 'fetch_json' }).then(function(token) {
                    recaptchaToken = token;
                    // console.log("reCAPTCHA token generated:", recaptchaToken);
                    loadData();
                }).catch(function(error) {
                    console.error("Error generating reCAPTCHA token:", error);
                });
            });
        } else {
            console.error("reCAPTCHA script not loaded correctly.");
        }
    }

    function loadData() {
        $.when(
            fetchJSON('resume.json', function(data) {
                jobs = data;
            }),
            fetchJSON('portfolio.json', function(data) {
                portfolioEntries = data;
            }),
            fetchJSON('contacts.json', function(data) {
                contacts = data;
            }),
            fetchJSON('about.json', function(data) {
                aboutContent = data;
            })
        ).then(function() {
            $('#mute-button').click(toggleMute);
            displayAbout();
            document.addEventListener('click', playAudio);
            $('#resume-tab').trigger('click');

        });
    }

    $('#resume-tab').click(function() {
        $('#content').html('<h1>Resume</h1><div id="resume-content"></div>');
        displayJobs();
    });

    $('#portfolio-tab').click(function() {
        $('#content').html('<h1>Portfolio</h1><div id="portfolio-content"></div>');
        displayPortfolio();
    });

    $('#github-tab').click(function() {
        $('#content').html(`
            <marquee scrollamount="4" style="background-color: red;">
                <font color="white" size="3">
                    <b>All hope abandon, þe who enter here!!!</b>
                </font>
            </marquee>
            <div id="github-content" style="text-align: center; margin-top: 20px;">
                <a href="https://github.com/QuiteLiterallyConnor" target="_blank">
                    <img src="https://www.analyticsvidhya.com/wp-content/uploads/2015/07/github_logo-1024x219.png" alt="GitHub" style="width: 200px;">
                </a>
            </div>
        `);
    });

    $('#contacts-tab').click(function() {
        displayContacts();
    });

    initializeRecaptcha();

    // Automatically open Resume tab on page load
});
