$(document).ready(function() {
    let allFaders = []; // Expects items with lowercase 'id', 'Name', 'Type' from /api/faders
    let currentFaderId = null; // Will store lowercase 'id'
    let socket = null;

    // Initial UI State
    $('#fader-slider').prop('disabled', true);
    $('#mute-button').prop('disabled', true);
    $('#fader-item-select').prop('disabled', true);

    // Helper function to update fader UI elements.
    // Expects faderData to have lowercase 'id'.
    function updateFaderUI(faderData) {
        if (!faderData) {
            $('#selected-fader-name').text('No Fader Selected');
            $('#fader-slider').val(0).prop('disabled', true);
            $('#fader-db-value').text('0.0 dB');
            $('#mute-button').text('Mute').removeClass('btn-success').addClass('btn-danger').prop('disabled', true);
            currentFaderId = null;
            return;
        }

        currentFaderId = faderData.id; // Expects lowercase 'id'
        $('#selected-fader-name').text(`${faderData.Name} (${faderData.id})`); // Expects lowercase 'id'
        $('#fader-slider').val(faderData.Level).prop('disabled', false);
        $('#fader-db-value').text(`${parseFloat(faderData.Level).toFixed(1)} dB`);

        if (faderData.Muted) {
            $('#mute-button').text('Unmute').removeClass('btn-danger').addClass('btn-success').prop('disabled', false);
        } else {
            $('#mute-button').text('Mute').removeClass('btn-success').addClass('btn-danger').prop('disabled', false);
        }
    }

    // 1. Fetch All Faders and Populate Type Selector
    // Server's /api/faders provides { "id": "...", "name": "...", "type": "..." }
    $.ajax({
        url: '/api/faders',
        method: 'GET',
        dataType: 'json',
        success: function(data) {
            allFaders = data; // `data` is an array of objects with lowercase 'id'
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
            // Items in allFaders have lowercase 'id'
            filteredFaders.sort((a, b) => a.id.localeCompare(b.id)); 
            filteredFaders.forEach(function(fader) { // fader here has lowercase 'id'
                $('#fader-item-select').append($('<option>', {
                    value: fader.id, // Use lowercase 'id'
                    text: `${fader.Name} (${fader.id})` // Use lowercase 'id'
                }));
            });
            $('#fader-item-select').prop('disabled', false);
        }
    });

    // 3. Fetch and Display Fader Details (On Item Change)
    $('#fader-item-select').on('change', function() {
        const faderIdFromSelect = $(this).val(); // This is lowercase 'id'
        if (!faderIdFromSelect || faderIdFromSelect === "Select Item...") {
            updateFaderUI(null);
            return;
        }

        // Server's /api/faders/{id} provides { "ID": "...", "Name": "...", ... } (uppercase ID)
        $.ajax({
            url: `/api/faders/${faderIdFromSelect}`,
            method: 'GET',
            dataType: 'json',
            success: function(faderDataFromServer) {
                // Normalize to lowercase 'id' before passing to updateFaderUI
                const normalizedFaderData = {
                    id: faderDataFromServer.ID, // Map uppercase ID to lowercase id
                    Name: faderDataFromServer.Name,
                    Type: faderDataFromServer.Type,
                    Level: faderDataFromServer.Level,
                    Muted: faderDataFromServer.Muted
                };
                updateFaderUI(normalizedFaderData);
            },
            error: function(jqXHR, textStatus, errorThrown) {
                console.error(`Error fetching fader ${faderIdFromSelect}:`, textStatus, errorThrown);
                updateFaderUI(null);
            }
        });
    });

    // 4. Update Fader Level (On Slider Input)
    $('#fader-slider').on('input', function() {
        const newLevel = parseFloat($(this).val());
        $('#fader-db-value').text(`${newLevel.toFixed(1)} dB`);

        if (currentFaderId) { // currentFaderId is lowercase 'id'
            $.ajax({
                url: `/api/faders/${currentFaderId}`, // API path uses the ID
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ level: newLevel }),
                dataType: 'json',
                success: function(updatedFaderFromServer) {
                    // POST response also has uppercase ID.
                    // If we were to use this for UI update, normalization would be needed.
                    // console.log(`Fader ${currentFaderId} level updated to ${updatedFaderFromServer.Level}`);
                    // The WebSocket message will trigger the UI update for the current fader.
                },
                error: function(jqXHR, textStatus, errorThrown) {
                    console.error(`Error updating fader ${currentFaderId} level:`, textStatus, errorThrown);
                }
            });
        }
    });

    // 5. Toggle Mute (On Mute Button Click)
    $('#mute-button').on('click', function() {
        if (!currentFaderId) return; // currentFaderId is lowercase 'id'

        const isCurrentlyMuted = $(this).hasClass('btn-success'); 
        const newMuteState = !isCurrentlyMuted;

        $.ajax({
            url: `/api/faders/${currentFaderId}`, // API path uses the ID
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ muted: newMuteState }),
            dataType: 'json',
            success: function(updatedFaderFromServer) {
                // POST response also has uppercase ID.
                // Normalize for updateFaderUI if not relying on WebSocket for this update.
                // The WebSocket message should ideally be the source of truth for UI updates.
                const normalizedFaderData = {
                    id: updatedFaderFromServer.ID,
                    Name: updatedFaderFromServer.Name,
                    Type: updatedFaderFromServer.Type,
                    Level: updatedFaderFromServer.Level,
                    Muted: updatedFaderFromServer.Muted
                };
                // updateFaderUI(normalizedFaderData); // Can be redundant if WebSocket updates are fast
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
            console.log("WebSocket raw message received:", event.data); // Added detailed log
            try {
                // WebSocket data comes from core.Fader, which has uppercase ID.
                const faderDataFromServer = JSON.parse(event.data);
                
                // Normalize to lowercase 'id' for client-side consistency
                const faderData = {
                    id: faderDataFromServer.ID, // Map uppercase ID to lowercase id
                    Name: faderDataFromServer.Name,
                    Type: faderDataFromServer.Type,
                    Level: faderDataFromServer.Level,
                    Muted: faderDataFromServer.Muted
                };

                if (faderData && faderData.id) { // Check lowercase 'id'
                    // Update item in allFaders array (which uses lowercase 'id')
                    let existingFaderIndex = allFaders.findIndex(f => f.id === faderData.id);
                    if (existingFaderIndex !== -1) {
                        allFaders[existingFaderIndex] = {
                            ...allFaders[existingFaderIndex], 
                            ...faderData 
                        };
                        // console.log("Updated fader in allFaders:", faderData.id);
                    }
                    
                    if (faderData.id === currentFaderId) {
                        console.log("Updating current fader UI for ID (WebSocket):", faderData.id);
                        updateFaderUI(faderData); // Pass normalized data
                    }
                } else {
                    console.warn("Received invalid fader data from WebSocket:", faderDataFromServer);
                }
            } catch (e) {
                console.error("WebSocket: Error parsing JSON message:", e, "Raw data:", event.data); // Added detailed log
                return; // Stop processing this message if it's unparseable
            }
        };

        socket.onerror = function(event) {
            console.error("WebSocket error event:", event); // Log full event
        };

        socket.onclose = function(event) {
            console.log("WebSocket connection closed. Code:", event.code, "Reason:", event.reason, "wasClean:", event.wasClean); // Added wasClean
            console.log("Attempting to reconnect WebSocket in 5 seconds...");
            setTimeout(connectWs, 5000);
        };
    }

    connectWs();
});
