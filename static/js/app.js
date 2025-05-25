$(document).ready(function() {
    let allFaders = []; // Expects items with lowercase 'id', 'name', 'type' from /api/faders
    let currentFaderId = null; // Will store lowercase 'id'
    let socket = null;

    // Define Type Sort Order
    const faderTypeSortOrder = {
        "Channel": 1,
        "Bus": 2,
        "Matrix": 3,
        "DCA": 4,
        "Master": 5
    };

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
        $('#selected-fader-name').text(`${faderData.Name} (${faderData.id})`); // Expects faderData.Name (uppercase N) from normalized object
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
            allFaders = data; // `data` is an array of objects with lowercase 'id', 'name', 'type'
            let faderTypes = ['All'];
            let types = new Set(allFaders.map(fader => fader.type).filter(type => type && type.trim() !== ""));
            
            // Sort types according to faderTypeSortOrder for the dropdown
            let sortedUniqueTypes = Array.from(types).sort((a, b) => {
                const orderA = faderTypeSortOrder[a] || 999;
                const orderB = faderTypeSortOrder[b] || 999;
                if (orderA !== orderB) {
                    return orderA - orderB;
                }
                return a.localeCompare(b); // Fallback sort for types with same order or not in map
            });

            sortedUniqueTypes.forEach(type => faderTypes.push(type));
            
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
        const $faderItemSelect = $('#fader-item-select');
        $faderItemSelect.empty().append($('<option>', {
            selected: true,
            disabled: true,
            text: 'Select Item...'
        })).prop('disabled', true);
        updateFaderUI(null); 

        if (!selectedType || selectedType === "Select Type...") {
            return;
        }

        let itemsAdded = 0;

        if (selectedType === 'All') {
            const groupedFaders = {};
            allFaders.forEach(fader => { // allFaders items have lowercase 'type'
                if (!groupedFaders[fader.type]) {
                    groupedFaders[fader.type] = [];
                }
                groupedFaders[fader.type].push(fader);
            });

            let sortedGroupTypes = Object.keys(groupedFaders).sort((a, b) => {
                const orderA = faderTypeSortOrder[a] || 999;
                const orderB = faderTypeSortOrder[b] || 999;
                if (orderA !== orderB) {
                    return orderA - orderB;
                }
                return a.localeCompare(b); // Fallback for types with same order or not in map
            });

            sortedGroupTypes.forEach(groupType => {
                const $optgroup = $('<optgroup>').attr('label', groupType);
                groupedFaders[groupType].sort((a, b) => a.id.localeCompare(b.id)); // Sort faders by id within the group
                
                groupedFaders[groupType].forEach(function(fader) { // fader items have lowercase 'id' and 'name'
                    $optgroup.append($('<option>', {
                        value: fader.id, 
                        text: `${fader.name} (${fader.id})` 
                    }));
                    itemsAdded++;
                });
                $faderItemSelect.append($optgroup);
            });

        } else {
            const filteredFadersForType = allFaders.filter(fader => fader.type === selectedType); // fader.type is lowercase

            if (filteredFadersForType.length > 0) {
                filteredFadersForType.sort((a, b) => a.id.localeCompare(b.id)); // Sort by id
                
                const $optgroup = $('<optgroup>').attr('label', selectedType);
                filteredFadersForType.forEach(function(fader) { // fader items have lowercase 'id' and 'name'
                    $optgroup.append($('<option>', {
                        value: fader.id,
                        text: `${fader.name} (${fader.id})`
                    }));
                    itemsAdded++;
                });
                $faderItemSelect.append($optgroup);
            }
        }

        if (itemsAdded > 0) {
            $faderItemSelect.prop('disabled', false);
        }
    });

    // 3. Fetch and Display Fader Details (On Item Change)
    $('#fader-item-select').on('change', function() {
        const faderIdFromSelect = $(this).val(); 
        if (!faderIdFromSelect || faderIdFromSelect === "Select Item...") {
            updateFaderUI(null);
            return;
        }

        $.ajax({
            url: `/api/faders/${faderIdFromSelect}`,
            method: 'GET',
            dataType: 'json',
            success: function(faderDataFromServer) {
                const normalizedFaderData = {
                    id: faderDataFromServer.ID, 
                    Name: faderDataFromServer.Name, // Keep Name, Type, Level, Muted as is from full object
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

        if (currentFaderId) { 
            $.ajax({
                url: `/api/faders/${currentFaderId}`, 
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ level: newLevel }),
                dataType: 'json',
                success: function(updatedFaderFromServer) {
                    // WebSocket message will trigger the UI update.
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
            success: function(updatedFaderFromServer) {
                // WebSocket message should ideally be the source of truth for UI updates.
            },
            error: function(jqXHR, textStatus, errorThrown) {
                console.error(`Error updating fader ${currentFaderId} mute state:`, textStatus, errorThrown);
            }
        });
    });

    // 6. Reset Fader Level to 0dB on Double-Click
    $('#fader-slider').on('dblclick', function(event) { // Added event parameter
        event.preventDefault(); // Prevent default double-click behavior

        if (currentFaderId) {
            console.log("Fader slider double-clicked for ID:", currentFaderId, "- resetting to 0dB.");
            const newLevel = 0.0;
            $(this).val(newLevel); // Set the slider's value to 0
            $('#fader-db-value').text(newLevel.toFixed(1) + ' dB'); // Directly update dB display

            // Directly send AJAX POST request
            $.ajax({
                url: `/api/faders/${currentFaderId}`,
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ level: newLevel }),
                dataType: 'json',
                success: function(updatedFaderFromServer) {
                    console.log(`Fader ${currentFaderId} level successfully reset to 0dB via dblclick.`);
                    // WebSocket update should handle UI consistency if other clients are involved
                    // or if server modifies the value further.
                    // For immediate feedback, current UI update is already done.
                },
                error: function(jqXHR, textStatus, errorThrown) {
                    console.error(`Error resetting fader ${currentFaderId} level to 0dB:`, textStatus, errorThrown);
                    // Consider reverting UI or notifying user if the update fails
                }
            });
        } else {
            console.log("Fader slider double-clicked, but no fader selected.");
        }
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
            console.log("WebSocket raw message received:", event.data); 
            try {
                const faderDataFromServer = JSON.parse(event.data);
                
                const faderData = { // Normalized data with lowercase 'id'
                    id: faderDataFromServer.ID, 
                    Name: faderDataFromServer.Name, // Keep Name, Type, Level, Muted as is from full object
                    Type: faderDataFromServer.Type,
                    Level: faderDataFromServer.Level,
                    Muted: faderDataFromServer.Muted
                };

                if (faderData && faderData.id) { 
                    let existingFaderIndex = allFaders.findIndex(f => f.id === faderData.id);
                    if (existingFaderIndex !== -1) {
                        // Update the allFaders item. Note: allFaders items from /api/faders
                        // only have id, name, type. WebSocket gives full object.
                        // We merge to preserve original list structure if needed, but update with new full data.
                        allFaders[existingFaderIndex] = {
                            ...allFaders[existingFaderIndex], // existing item has {id, name, type}
                            Name: faderData.Name, // Update name if it changed
                            Type: faderData.Type, // Update type if it changed
                            // Add Level and Muted if we want allFaders to store full state
                            // For now, allFaders list items are just {id, name, type}
                            // So, we might only update what's there:
                            // allFaders[existingFaderIndex].name = faderData.Name;
                            // allFaders[existingFaderIndex].type = faderData.Type;
                            // Or, if we want allFaders to be a cache of full objects:
                             ...faderData
                        };
                    }
                    
                    if (faderData.id === currentFaderId) {
                        console.log("Updating current fader UI for ID (WebSocket):", faderData.id);
                        updateFaderUI(faderData); 
                    }
                } else {
                    console.warn("Received invalid fader data from WebSocket:", faderDataFromServer);
                }
            } catch (e) {
                console.error("WebSocket: Error parsing JSON message:", e, "Raw data:", event.data); 
                return; 
            }
        };

        socket.onerror = function(event) {
            console.error("WebSocket error event:", event); 
        };

        socket.onclose = function(event) {
            console.log("WebSocket connection closed. Code:", event.code, "Reason:", event.reason, "wasClean:", event.wasClean); 
            console.log("Attempting to reconnect WebSocket in 5 seconds...");
            setTimeout(connectWs, 5000);
        };
    }

    connectWs();
});
