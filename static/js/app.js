$(document).ready(function() {
    let allFaders = [];
    let currentFaderId = null;
    let socket = null; // Declare socket variable in a broader scope

    // Initial UI State
    $('#fader-slider').prop('disabled', true);
    $('#mute-button').prop('disabled', true);
    $('#fader-item-select').prop('disabled', true);

    // Helper function to update fader UI elements
    function updateFaderUI(faderData) {
        if (!faderData) {
            $('#selected-fader-name').text('No Fader Selected');
            $('#fader-slider').val(0).prop('disabled', true);
            $('#fader-db-value').text('0.0 dB');
            $('#mute-button').text('Mute').removeClass('btn-success').addClass('btn-danger').prop('disabled', true);
            currentFaderId = null;
            return;
        }

        currentFaderId = faderData.ID; // Update currentFaderId when a fader is displayed
        $('#selected-fader-name').text(`${faderData.Name} (${faderData.ID})`);
        $('#fader-slider').val(faderData.Level).prop('disabled', false);
        $('#fader-db-value').text(`${parseFloat(faderData.Level).toFixed(1)} dB`);

        if (faderData.Muted) {
            $('#mute-button').text('Unmute').removeClass('btn-danger').addClass('btn-success').prop('disabled', false);
        } else {
            $('#mute-button').text('Mute').removeClass('btn-success').addClass('btn-danger').prop('disabled', false);
        }
    }

    // 1. Fetch All Faders and Populate Type Selector
    $.ajax({
        url: '/api/faders',
        method: 'GET',
        dataType: 'json',
        success: function(data) {
            allFaders = data; // Initialize allFaders
            let faderTypes = ['All'];
            let types = new Set(allFaders.map(fader => fader.Type));
            types.forEach(type => faderTypes.push(type));
            
            $('#fader-type-select').empty().append($('<option>', {
                selected: true,
                disabled: true,
                text: 'Select Type...'
            }));

            faderTypes.forEach(function(type) {
                $('#fader-type-select').append($('<option>', {
                    value: type,
                    text: type
                }));
            });
        },
        error: function(jqXHR, textStatus, errorThrown) {
            console.error('Error fetching faders:', textStatus, errorThrown);
        }
    });

    // 2. Populate Fader Item Selector (On Type Change)
    $('#fader-type-select').on('change', function() {
        const selectedType = $(this).val();
        $('#fader-item-select').empty().append($('<option>', {
            selected: true,
            disabled: true,
            text: 'Select Item...'
        })).prop('disabled', true);
        updateFaderUI(null); 

        if (!selectedType || selectedType === "Select Type...") {
            return;
        }

        const filteredFaders = (selectedType === 'All') 
            ? allFaders 
            : allFaders.filter(fader => fader.Type === selectedType);

        if (filteredFaders.length > 0) {
            filteredFaders.sort((a, b) => a.ID.localeCompare(b.ID));
            filteredFaders.forEach(function(fader) {
                $('#fader-item-select').append($('<option>', {
                    value: fader.ID,
                    text: `${fader.Name} (${fader.ID})` 
                }));
            });
            $('#fader-item-select').prop('disabled', false);
        }
    });

    // 3. Fetch and Display Fader Details (On Item Change)
    $('#fader-item-select').on('change', function() {
        const faderId = $(this).val();
        if (!faderId || faderId === "Select Item...") {
            updateFaderUI(null);
            return;
        }

        $.ajax({
            url: `/api/faders/${faderId}`,
            method: 'GET',
            dataType: 'json',
            success: function(faderData) {
                updateFaderUI(faderData);
            },
            error: function(jqXHR, textStatus, errorThrown) {
                console.error(`Error fetching fader ${faderId}:`, textStatus, errorThrown);
                updateFaderUI(null);
            }
        });
    });

    // 4. Update Fader Level (On Slider Input)
    $('#fader-slider').on('input', function() {
        const newLevel = parseFloat($(this).val());
        $('#fader-db-value').text(`${newLevel.toFixed(1)} dB`);

        if (currentFaderId) {
            $.ajax({
                url: `/api/faders/${currentFaderId}`,
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ level: newLevel }),
                dataType: 'json',
                success: function(updatedFader) {
                    // console.log(`Fader ${currentFaderId} level updated to ${updatedFader.Level}`);
                },
                error: function(jqXHR, textStatus, errorThrown) {
                    console.error(`Error updating fader ${currentFaderId} level:`, textStatus, errorThrown);
                }
            });
        }
    });

    // 5. Toggle Mute (On Mute Button Click)
    $('#mute-button').on('click', function() {
        if (!currentFaderId) return;

        const isCurrentlyMuted = $(this).hasClass('btn-success'); 
        const newMuteState = !isCurrentlyMuted;

        $.ajax({
            url: `/api/faders/${currentFaderId}`,
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ muted: newMuteState }),
            dataType: 'json',
            success: function(updatedFader) {
                updateFaderUI(updatedFader); 
            },
            error: function(jqXHR, textStatus, errorThrown) {
                console.error(`Error updating fader ${currentFaderId} mute state:`, textStatus, errorThrown);
            }
        });
    });

    // --- WebSocket Implementation ---
    function connectWs() {
        const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
        
        console.log("Attempting to connect to WebSocket:", wsUrl);
        socket = new WebSocket(wsUrl);

        socket.onopen = function(event) {
            console.log("WebSocket connection established.");
        };

        socket.onmessage = function(event) {
            console.log("WebSocket message received:", event.data);
            try {
                const faderData = JSON.parse(event.data);

                if (faderData && faderData.ID) {
                    // Optional Enhancement: Update allFaders array
                    let existingFader = allFaders.find(f => f.ID === faderData.ID);
                    if (existingFader) {
                        Object.assign(existingFader, faderData);
                        // console.log("Updated fader in allFaders:", faderData.ID);
                    } else {
                        // If fader is not in allFaders (e.g. dynamically added on server),
                        // you might want to add it or handle it differently.
                        // For now, we just log this case.
                        // console.log("Received update for fader not initially in allFaders:", faderData.ID);
                    }
                    
                    // Update UI if the received fader data is for the currently selected fader
                    if (faderData.ID === currentFaderId) {
                        console.log("Updating current fader UI for ID:", faderData.ID);
                        updateFaderUI(faderData);
                    }
                } else {
                    console.warn("Received invalid fader data from WebSocket:", faderData);
                }
            } catch (e) {
                console.error("Error parsing WebSocket message JSON:", e);
            }
        };

        socket.onerror = function(event) {
            console.error("WebSocket error:", event);
        };

        socket.onclose = function(event) {
            console.log("WebSocket connection closed. Code:", event.code, "Reason:", event.reason);
            console.log("Attempting to reconnect WebSocket in 5 seconds...");
            setTimeout(connectWs, 5000); // Attempt to reconnect
        };
    }

    // Call connectWs to establish the WebSocket connection after initial setup
    connectWs();
});
